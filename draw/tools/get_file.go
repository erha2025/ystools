package tools

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
)

func GetImgInfo(path string) (int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		return 0, 0, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Printf("解码图片失败: %v\n", err)
		return 0, 0, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// fmt.Printf("图片宽度: %d\n", width)
	// fmt.Printf("图片高度: %d\n", height)
	return width, height, nil
}

// DrawRedLine 在指定图片上画一条红色水平线
// inputPath: 输入图片路径
// outputPath: 输出图片路径
// y: 红线的y坐标
// 返回: 错误信息
func DrawRedLine(inputPath, outputPath string, y int) error {
	return DrawRedLines(inputPath, outputPath, []int{y})
}

// DrawRedLines 在指定图片上画多条红色水平线
// inputPath: 输入图片路径
// outputPath: 输出图片路径
// yPositions: 各条红线的y坐标列表
// 返回: 错误信息
func DrawRedLines(inputPath, outputPath string, yPositions []int) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("解码图片失败: %v", err)
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	red := color.RGBA{255, 0, 0, 255}
	for _, y := range yPositions {
		for x := 0; x < bounds.Dx(); x++ {
			rgba.Set(x, y, red)
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outFile.Close()

	switch format {
	case "jpeg":
		return jpeg.Encode(outFile, rgba, &jpeg.Options{Quality: 95})
	case "png":
		return png.Encode(outFile, rgba)
	default:
		return fmt.Errorf("不支持的图片格式: %s", format)
	}
}

// DrawRedVerticalLines 在指定图片上画多条红色垂直线
// inputPath: 输入图片路径
// outputPath: 输出图片路径
// xPositions: 各条垂直线的x坐标列表
// 返回: 错误信息
func DrawRedVerticalLines(inputPath, outputPath string, xPositions []int) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("解码图片失败: %v", err)
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	red := color.RGBA{255, 0, 0, 255}
	for _, x := range xPositions {
		for y := 0; y < bounds.Dy(); y++ {
			rgba.Set(x, y, red)
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outFile.Close()

	switch format {
	case "jpeg":
		return jpeg.Encode(outFile, rgba, &jpeg.Options{Quality: 95})
	case "png":
		return png.Encode(outFile, rgba)
	default:
		return fmt.Errorf("不支持的图片格式: %s", format)
	}
}

// DrawGrid 在指定图片上画网格（水平线和垂直线）
// inputPath: 输入图片路径
// outputPath: 输出图片路径
// yPositions: 水平线的y坐标列表
// xPositions: 垂直线的x坐标列表
// 返回: 错误信息
func DrawGrid(inputPath, outputPath string, yPositions, xPositions []int) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("解码图片失败: %v", err)
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	red := color.RGBA{255, 0, 0, 255}

	// 画水平线
	for _, y := range yPositions {
		for x := 0; x < bounds.Dx(); x++ {
			rgba.Set(x, y, red)
		}
	}

	// 画垂直线
	for _, x := range xPositions {
		for y := 0; y < bounds.Dy(); y++ {
			rgba.Set(x, y, red)
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outFile.Close()

	switch format {
	case "jpeg":
		return jpeg.Encode(outFile, rgba, &jpeg.Options{Quality: 95})
	case "png":
		return png.Encode(outFile, rgba)
	default:
		return fmt.Errorf("不支持的图片格式: %s", format)
	}
}
