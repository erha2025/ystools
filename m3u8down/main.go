package main

import (
	"crypto/aes"
	"crypto/cipher"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// 固定配置：4并发下载
const concurrentNum = 4

// 进度统计（并发安全）
type Progress struct {
	total   int
	done    int
	failed  int
	success int
	mu      sync.Mutex
}

func NewProgress(total int) *Progress {
	return &Progress{total: total}
}

// 完成一个下载，更新进度
func (p *Progress) Done(success bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done++
	if success {
		p.success++
	} else {
		p.failed++
	}
	// 单行刷新进度（不刷屏）
	fmt.Printf("\r📥 下载进度：%d/%d [成功: %d, 失败: %d] (%.2f%%)",
		p.done, p.total, p.success, p.failed, float64(p.done)/float64(p.total)*100)
}

// DownloadM3U8 核心函数：m3u8链接 + 保存文件名
func DownloadM3U8(m3u8Url, saveName string) error {
	// 开始计时
	startTime := time.Now()
	fmt.Printf("🚀 开始下载：%s\n", saveName)

	// 创建临时目录
	tempDir := filepath.Join("./ts_temp", strings.TrimSuffix(saveName, ".mp4"))
	_ = os.MkdirAll(tempDir, 0755)

	// 确保退出时清理
	defer func() {
		fmt.Println("\n🧹 清理临时文件...")
		_ = os.RemoveAll(tempDir)
	}()

	client := &http.Client{Timeout: 60 * time.Second}

	// 1. 解析m3u8，获取ts列表和密钥
	tsList, key, err := parseM3U8(client, m3u8Url)
	if err != nil {
		return fmt.Errorf("解析m3u8失败: %w", err)
	}
	total := len(tsList)
	progress := NewProgress(total)
	fmt.Printf("📦 解析完成，总分片：%d 个\n", total)

	// 2. 并发下载（控制4协程）+ 重试机制
	tsPaths := make([]string, total)
	downloadSuccess := make([]bool, total)
	var downloadMu sync.Mutex

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrentNum) // 信号量限流

	for idx, tsUrl := range tsList {
		wg.Add(1)
		sem <- struct{}{}

		go func(index int, url string) {
			defer wg.Done()
			defer func() { <-sem }()

			tsPath := filepath.Join(tempDir, fmt.Sprintf("%06d.ts", index))
			tsPaths[index] = tsPath

			// 重试3次
			var lastErr error
			for retry := 0; retry < 3; retry++ {
				err := downloadTs(client, url, key, tsPath)
				if err == nil {
					// 验证文件是否下载成功
					info, statErr := os.Stat(tsPath)
					if statErr == nil && info.Size() > 0 {
						downloadMu.Lock()
						downloadSuccess[index] = true
						downloadMu.Unlock()
						progress.Done(true)
						return
					}
				}
				lastErr = err
				time.Sleep(500 * time.Millisecond) // 重试间隔
			}

			fmt.Printf("\n⚠️ 下载失败 [%d]: %v (URL: %s)\n", index, lastErr, url)
			progress.Done(false)
		}(idx, tsUrl)
	}

	// 等待所有下载完成
	wg.Wait()
	fmt.Println() // 换行

	// 3. 检查下载结果
	successCount := 0
	for _, ok := range downloadSuccess {
		if ok {
			successCount++
		}
	}

	if successCount == 0 {
		return fmt.Errorf("所有分片下载失败，无法合并")
	}

	if successCount < total {
		fmt.Printf("⚠️ 部分下载失败：成功 %d/%d，将只合并成功的分片\n", successCount, total)
	}

	// 4. 合并为MP4
	fmt.Println("🔧 开始合并视频...")
	if err := mergeTsToMP4(tempDir, saveName, downloadSuccess); err != nil {
		return fmt.Errorf("合并失败: %w", err)
	}

	// 统计总耗时
	cost := time.Since(startTime).Round(time.Second)
	fmt.Printf("✅ 下载完成！总耗时：%v\n文件保存至：%s\n", cost, saveName)
	return nil
}

// parseM3U8 解析m3u8索引
func parseM3U8(client *http.Client, url string) ([]string, []byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	content := string(body)
	baseUrl := url[:strings.LastIndex(url, "/")+1]

	var tsList []string
	var key []byte

	// 提取AES密钥
	keyRegex := regexp.MustCompile(`URI="([^"]+)"`)
	if match := keyRegex.FindStringSubmatch(content); match != nil {
		keyUrl := match[1]
		if !strings.HasPrefix(keyUrl, "http") {
			keyUrl = baseUrl + keyUrl
		}
		fmt.Printf("🔑 获取密钥: %s\n", keyUrl)
		kResp, err := client.Get(keyUrl)
		if err != nil {
			fmt.Printf("⚠️ 密钥获取失败: %v\n", err)
		} else {
			defer kResp.Body.Close()
			key, _ = io.ReadAll(kResp.Body)
			fmt.Printf("🔑 密钥长度: %d bytes\n", len(key))
		}
	}

	// 提取所有TS链接
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(line, "http") {
			line = baseUrl + line
		}
		tsList = append(tsList, line)
	}

	return tsList, key, nil
}

// downloadTs 下载并解密TS
func downloadTs(client *http.Client, url string, key []byte, savePath string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if len(data) < 188 { // TS包最小大小
		return fmt.Errorf("数据太小，不是有效的TS文件")
	}

	// AES-128解密
	if len(key) == 16 {
		data = aesDecrypt(data, key)
	}

	return os.WriteFile(savePath, data, 0644)
}

// aesDecrypt AES-128-CBC 解密（标准m3u8加密）
func aesDecrypt(data, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		return data // 解密失败返回原数据
	}
	iv := make([]byte, aes.BlockSize) // 默认IV全0
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(data, data)
	return pkcs7Unpad(data)
}

// pkcs7Unpad 去除填充
func pkcs7Unpad(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return data
	}
	unpad := int(data[length-1])
	if unpad > length || unpad == 0 {
		return data
	}
	return data[:length-unpad]
}

// mergeTsToMP4 调用ffmpeg合并
func mergeTsToMP4(tempDir, outName string, success []bool) error {
	listPath := filepath.Join(tempDir, "filelist.txt")

	// 先删除旧文件
	_ = os.Remove(listPath)

	f, err := os.Create(listPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// 排序并只添加成功的文件
	files, _ := filepath.Glob(filepath.Join(tempDir, "*.ts"))
	sort.Strings(files)

	successCount := 0
	for i, ts := range files {
		if i < len(success) && success[i] {
			_, err := f.WriteString("file '" + filepath.Base(ts) + "'\n")
			if err != nil {
				continue
			}
			successCount++
		}
	}

	if successCount == 0 {
		return fmt.Errorf("没有可合并的文件")
	}

	fmt.Printf("📝 合并文件列表: %d 个文件\n", successCount)

	// ffmpeg合并命令 - 使用绝对路径
	absListPath, _ := filepath.Abs(listPath)
	absOutName, _ := filepath.Abs(outName)
	absTempDir, _ := filepath.Abs(tempDir)

	// 确保输出目录存在
	outDir := filepath.Dir(absOutName)
	if outDir != "." && outDir != "" {
		_ = os.MkdirAll(outDir, 0755)
	}

	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", absListPath,
		"-c", "copy",
		absOutName,
	)
	cmd.Dir = absTempDir // 在tempDir目录下执行

	output, err := cmd.CombinedOutput()
	if err != nil {
		// 打印ffmpeg错误输出
		fmt.Printf("❌ ffmpeg错误输出:\n%s\n", string(output))
		return fmt.Errorf("ffmpeg执行失败: %w", err)
	}

	return nil
}

// ===================== 主函数：使用flag接收参数 =====================
func main() {
	// 使用flag接收命令行参数
	url := flag.String("url", "", "m3u8视频URL地址")
	name := flag.String("name", "output.mp4", "输出文件名")
	flag.Parse()

	// 检查必要参数
	if *url == "" {
		fmt.Println("❌ 请提供m3u8 URL地址")
		fmt.Println("\n📖 使用方法:")
		fmt.Println("  go run main.go -url <m3u8_url> -name <output_name>")
		fmt.Println("\n📝 示例:")
		fmt.Println("  go run main.go -url https://example.com/video.m3u8 -name video.mp4")
		flag.Usage()
		return
	}

	// 执行下载
	fmt.Printf("📹 视频URL: %s\n", *url)
	fmt.Printf("💾 输出文件: %s\n", *name)

	if err := DownloadM3U8(*url, *name); err != nil {
		fmt.Printf("❌ 下载失败：%v\n", err)
	}
}
