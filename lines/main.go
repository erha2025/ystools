package main

import (
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"ystools/lines/tools"
)

type ImageTask struct {
	inputPath      string
	outputPath     string
	filename       string
	outputFilename string
}

type WorkerPool struct {
	tasks       chan ImageTask
	wg          sync.WaitGroup
	workerCount int
	processed   atomic.Int64
}

func NewWorkerPool(workerCount int) *WorkerPool {
	pool := &WorkerPool{
		tasks:       make(chan ImageTask, 100),
		workerCount: workerCount,
	}

	for i := 0; i < workerCount; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	return pool
}

const (
	gridSize = 3 //格子大小
)

func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	for task := range p.tasks {
		width, height, err := tools.GetImgInfo(task.inputPath)
		if err != nil {
			fmt.Printf("获取图片信息失败 [%s]: %v\n", task.filename, err)
			continue
		}

		var sw, hw, min int
		sw = width / gridSize
		hw = height / gridSize
		if sw > hw {
			min = hw
		} else {
			min = sw
		}

		var yPositions []int
		for y := min; y < height; y += min {
			yPositions = append(yPositions, y)
		}

		var xPositions []int
		for x := min; x < width; x += min {
			xPositions = append(xPositions, x)
		}

		err = tools.DrawGrid(task.inputPath, task.outputPath, yPositions, xPositions)
		if err != nil {
			fmt.Printf("画线失败 [%s]: %v\n", task.filename, err)
			continue
		}

		fmt.Printf("Worker-%d: 已处理: %s -> %s (min=%d, 水平线=%d, 垂直线=%d)\n",
			id, task.filename, task.outputFilename, min, len(yPositions), len(xPositions))
		p.processed.Add(1)
	}
}

func (p *WorkerPool) Submit(task ImageTask) {
	p.tasks <- task
}

func (p *WorkerPool) Close() {
	close(p.tasks)
	p.wg.Wait()
}

func (p *WorkerPool) GetProcessedCount() int64 {
	return p.processed.Load()
}

func main() {
	folder := flag.String("folder", "", "图片目录路径")
	del := flag.Bool("del", false, "处理完成后是否删除原文件")
	flag.Parse()

	if *folder == "" {
		fmt.Println("请使用 -folder 参数指定图片目录")
		flag.Usage()
		return
	}

	files, err := os.ReadDir(*folder)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		return
	}

	pool := NewWorkerPool(3)

	var processedFiles []string
	taskCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			continue
		}

		inputPath := filepath.Join(*folder, filename)

		baseName := strings.TrimSuffix(filename, ext)
		outputExt := "_lines.jpg"
		if ext == ".png" {
			outputExt = "_lines.png"
		}
		outputFilename := baseName + outputExt
		outputPath := filepath.Join(*folder, outputFilename)

		processedFiles = append(processedFiles, inputPath)

		task := ImageTask{
			inputPath:      inputPath,
			outputPath:     outputPath,
			filename:       filename,
			outputFilename: outputFilename,
		}

		pool.Submit(task)
		taskCount++
	}

	pool.Close()

	if *del && len(processedFiles) > 0 {
		fmt.Println("\n删除原文件...")
		for _, filePath := range processedFiles {
			err := os.Remove(filePath)
			if err != nil {
				fmt.Printf("删除失败 %s: %v\n", filepath.Base(filePath), err)
			} else {
				fmt.Printf("已删除: %s\n", filepath.Base(filePath))
			}
		}
	}

	fmt.Printf("\n处理完成！共处理 %d 张图片\n", pool.GetProcessedCount())
}

// go run main.go -folder "图片目录路径" [-del]
// 示例: go run main.go -folder "/Users/yangsen/Pictures" -del
