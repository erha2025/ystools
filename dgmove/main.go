package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// 定义命令行参数
	path := flag.String("src", "", "源文件夹路径")
	targetPath := flag.String("dst", "", "目标文件夹路径")

	// 解析命令行参数
	flag.Parse()

	// 检查参数是否提供
	if *path == "" || *targetPath == "" {
		fmt.Println("用法: dgmove -src <源文件夹路径> -dst <目标文件夹路径>")
		fmt.Println("")
		fmt.Println("示例:")
		fmt.Println("  dgmove -src /Users/yangsen/Downloads/pixiv -dst /Users/yangsen/Downloads/todo")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// 检查源路径是否存在
	if _, err := os.Stat(*path); os.IsNotExist(err) {
		fmt.Printf("错误: 源文件夹不存在: %s\n", *path)
		os.Exit(1)
	}

	fmt.Printf("开始移动文件...\n")
	fmt.Printf("源路径: %s\n", *path)
	fmt.Printf("目标路径: %s\n", *targetPath)

	MoveImages(*path, *targetPath)
}

func MoveImages(path, targetPath string) {
	// 确保目标文件夹存在
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		fmt.Printf("创建目标文件夹失败: %v\n", err)
		return
	}

	var count int = 1

	// 递归遍历文件夹
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过文件夹
		if info.IsDir() {
			return nil
		}

		// 获取文件扩展名
		ext := filepath.Ext(filePath)

		// 生成新文件名（00001格式）
		newFileName := fmt.Sprintf("%05d%s", count, ext)
		count++

		// 目标文件路径
		targetFilePath := filepath.Join(targetPath, newFileName)

		// 移动文件
		if err := os.Rename(filePath, targetFilePath); err != nil {
			fmt.Printf("移动文件失败 %s -> %s: %v\n", filePath, targetFilePath, err)
			return nil // 继续处理其他文件
		}

		fmt.Printf("移动成功: %s\n", newFileName)
		return nil
	})

	if err != nil {
		fmt.Printf("遍历文件夹失败: %v\n", err)
	}

	fmt.Printf("移动完成，共处理 %d 个文件\n", count-1)
}

//go run main.go -src <源文件夹路径> -dst <目标文件夹路径>
