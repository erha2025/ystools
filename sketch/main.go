package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
)

// 协程池配置
const poolSize = 4

// 处理结果
type Result struct {
	inputPath  string
	outputPath string
	success    bool
	err        error
}

// toLineDraft 图片转线稿核心函数
func toLineDraft(src image.Image) image.Image {
	// 1. 转为灰度图
	gray := imaging.Grayscale(src)

	// 2. 高斯模糊（参数：模糊半径，数值越大线稿越粗）
	blur := imaging.Blur(gray, 3.0)

	// 3. 模糊图反色
	invertBlur := imaging.Invert(blur)

	// 4. 颜色减淡混合（核心：灰度图 + 反色模糊图 = 线稿）
	bounds := gray.Bounds()
	lineDraft := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// 原始灰度值
			g, _, _, _ := gray.At(x, y).RGBA()
			// 反色模糊后的灰度值
			ib, _, _, _ := invertBlur.At(x, y).RGBA()

			// 颜色减淡公式
			var val uint32
			if ib == 0xFFFF {
				val = 0xFFFF
			} else {
				val = (g * 0xFFFF) / (0xFFFF - ib)
				if val > 0xFFFF {
					val = 0xFFFF
				}
			}

			lineDraft.SetGray(x, y, color.Gray{Y: uint8(val >> 8)})
		}
	}

	return lineDraft
}

// 处理单张图片
func processImage(inputPath string, outputDir string) Result {
	result := Result{
		inputPath: inputPath,
	}

	// 生成输出文件名: 24.jpg -> 24_sk.png
	ext := filepath.Ext(inputPath)
	name := strings.TrimSuffix(filepath.Base(inputPath), ext)
	// 输出为 PNG 格式，保存到指定的输出目录
	outputPath := filepath.Join(outputDir, name+"_sk.png")
	result.outputPath = outputPath

	// 打开图片
	img, err := imaging.Open(inputPath)
	if err != nil {
		result.err = fmt.Errorf("打开图片失败: %w", err)
		return result
	}

	// 转换为线稿
	lineImg := toLineDraft(img)

	// 保存为PNG
	file, err := os.Create(outputPath)
	if err != nil {
		result.err = fmt.Errorf("创建文件失败: %w", err)
		return result
	}
	defer file.Close()

	err = png.Encode(file, lineImg)
	if err != nil {
		result.err = fmt.Errorf("保存图片失败: %w", err)
		return result
	}

	result.success = true
	return result
}

// 获取目录下所有支持的图片文件
func getImageFiles(dir string) ([]string, error) {
	var files []string
	supportedExts := map[string]bool{".jpg": true, ".jpeg": true, ".webp": true, ".png": true}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(path))
		if supportedExts[ext] {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// 协程池处理图片
func processWithPool(files []string, outputDir string) {
	if len(files) == 0 {
		fmt.Println("未找到支持的图片文件 (jpg, webp, png)")
		return
	}

	fmt.Printf("找到 %d 个图片文件\n", len(files))
	fmt.Printf("输出目录: %s\n", outputDir)
	fmt.Printf("启动协程池，大小: %d\n\n", poolSize)

	// 创建任务通道
	tasks := make(chan string, len(files))
	// 创建结果通道
	results := make(chan Result, len(files))

	// 创建 WaitGroup
	var wg sync.WaitGroup

	// 启动指定数量的 worker
	for i := 0; i < poolSize; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			fmt.Printf("Worker %d 启动\n", workerID+1)

			for path := range tasks {
				fmt.Printf("Worker %d 处理: %s\n", workerID+1, filepath.Base(path))
				result := processImage(path, outputDir)
				results <- result
			}

			fmt.Printf("Worker %d 结束\n", workerID+1)
		}(i)
	}

	// 发送所有任务
	go func() {
		for _, file := range files {
			tasks <- file
		}
		close(tasks)
	}()

	// 等待所有 worker 结束
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	successCount := 0
	failCount := 0

	for result := range results {
		if result.success {
			successCount++
			fmt.Printf("✅ 成功: %s -> %s\n", filepath.Base(result.inputPath), filepath.Base(result.outputPath))
		} else {
			failCount++
			fmt.Printf("❌ 失败: %s - %v\n", filepath.Base(result.inputPath), result.err)
		}
	}

	// 打印统计
	fmt.Printf("\n========== 处理完成 ==========\n")
	fmt.Printf("✅ 成功: %d 个\n", successCount)
	fmt.Printf("❌ 失败: %d 个\n", failCount)
	fmt.Printf("总计: %d 个\n", len(files))
}

func main() {
	// 使用flag接收目录参数
	dir := flag.String("dir", "", "图片目录路径")
	flag.Parse()

	// 检查参数
	if *dir == "" {
		fmt.Println("❌ 请提供图片目录路径")
		fmt.Println("\n📖 使用方法:")
		fmt.Println("  go run main.go -dir <目录路径>")
		fmt.Println("\n📝 示例:")
		fmt.Println("  go run main.go -dir /Users/yangsen/Pictures")
		flag.Usage()
		return
	}

	// 检查目录是否存在
	info, err := os.Stat(*dir)
	if err != nil {
		fmt.Printf("❌ 目录不存在: %s\n", *dir)
		return
	}
	if !info.IsDir() {
		fmt.Printf("❌ 路径不是目录: %s\n", *dir)
		return
	}

	fmt.Printf("📁 图片目录: %s\n\n", *dir)

	// 创建输出目录：在输入目录的同级目录下创建 原文件夹名_sk 的文件夹
	dirName := filepath.Base(*dir)
	parentDir := filepath.Dir(*dir)
	outputDir := filepath.Join(parentDir, dirName+"_sk")

	// 创建输出目录
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		fmt.Printf("❌ 创建输出目录失败: %v\n", err)
		return
	}
	fmt.Printf("📂 输出目录: %s\n\n", outputDir)

	// 获取目录下所有图片文件
	files, err := getImageFiles(*dir)
	if err != nil {
		fmt.Printf("❌ 遍历目录失败: %v\n", err)
		return
	}

	// 使用协程池处理
	processWithPool(files, outputDir)
}

//  运行  ./sketch -dir /Users/yangsen/Downloads
