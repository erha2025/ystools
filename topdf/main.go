package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

func main() {
	folder := flag.String("folder", "", "图片目录路径")
	flag.Parse()

	if *folder == "" {
		fmt.Println("请使用 -folder 参数指定图片目录")
		flag.Usage()
		return
	}

	// 检查目录是否存在
	if _, err := os.Stat(*folder); os.IsNotExist(err) {
		fmt.Printf("目录不存在: %s\n", *folder)
		return
	}

	// 获取目标文件夹名称
	folderName := filepath.Base(*folder)

	files, err := os.ReadDir(*folder)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		return
	}

	// 收集图片文件
	var imageFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			continue
		}

		imageFiles = append(imageFiles, filepath.Join(*folder, filename))
	}

	if len(imageFiles) == 0 {
		fmt.Println("目录中没有找到图片文件")
		return
	}

	// 生成 PDF
	pdfFilename := folderName + ".pdf"
	err = generatePDF(imageFiles, pdfFilename)
	if err != nil {
		fmt.Printf("生成 PDF 失败: %v\n", err)
		return
	}

	fmt.Printf("\nPDF 生成完成！共处理 %d 张图片\n", len(imageFiles))
	fmt.Printf("PDF 文件保存在: %s\n", pdfFilename)
}

func generatePDF(imagePaths []string, outputPath string) error {
	pdf := gofpdf.New("", "", "", "")

	for _, imgPath := range imagePaths {
		// 获取图片信息
		file, err := os.Open(imgPath)
		if err != nil {
			return fmt.Errorf("打开图片失败 %s: %v", imgPath, err)
		}

		img, _, err := image.Decode(file)
		file.Close()
		if err != nil {
			return fmt.Errorf("解码图片失败 %s: %v", imgPath, err)
		}

		bounds := img.Bounds()
		width := float64(bounds.Dx())
		height := float64(bounds.Dy())

		// A4 尺寸（毫米）
		const a4Width = 210.0
		const a4Height = 297.0

		// 计算缩放比例，保持宽高比
		var scale float64
		var pageWidth, pageHeight float64

		if width > height {
			// 横向
			pageWidth = a4Height
			pageHeight = a4Width
			scale = pageWidth / width
			if height*scale > pageHeight {
				scale = pageHeight / height
			}
		} else {
			// 纵向
			pageWidth = a4Width
			pageHeight = a4Height
			scale = pageWidth / width
			if height*scale > pageHeight {
				scale = pageHeight / height
			}
		}

		// 添加新页面
		pdf.AddPageFormat("", gofpdf.SizeType{Wd: pageWidth, Ht: pageHeight})

		// 计算图片位置（居中）
		imgWidth := width * scale
		imgHeight := height * scale
		x := (pageWidth - imgWidth) / 2
		y := (pageHeight - imgHeight) / 2

		// 插入图片
		pdf.Image(imgPath, x, y, imgWidth, imgHeight, false, "", 0, "")
	}

	// 保存 PDF
	return pdf.OutputFileAndClose(outputPath)
}

// go run main.go -folder "图片目录路径"
