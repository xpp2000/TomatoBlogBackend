package storage

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/nickalie/go-webpbin"
)

// CompressToWebP 将图片流实时压缩为 WebP 格式
// 参数 input: 原始图片文件流
// 参数 quality: 压缩质量 (0-100，推荐 80)
// 返回: 压缩后的内存缓冲区，以及可能发生的错误
func CompressToWebP(input io.Reader, quality uint) (*bytes.Buffer, error) {
	// 创建一个内存缓冲区，用来接住 cwebp 吐出来的压缩后数据
	var outputBuffer bytes.Buffer

	// 极其优雅的链式调用，全程在内存中流转，不产生任何临时文件！
	err := webpbin.NewCWebP().
		Quality(quality).
		Input(input).
		Output(&outputBuffer).
		Run()

	if err != nil {
		return nil, fmt.Errorf("WebP 压缩失败: %w", err)
	}

	return &outputBuffer, nil
}

// SmartCompressToWebP 智能自适应 WebP 压缩
// 参数 file: 原始文件对象
// 参数 targetSizeKB: 期望的最大体积 (建议传 200)
// 返回: 压缩后的内存缓冲区
func SmartCompressToWebP(file multipart.File, targetSizeKB int64) (*bytes.Buffer, error) {
	// 1. 🌟 避坑指南：将流一次性读入内存，防止多次压缩时遇到 EOF
	// 因为前面我们已经做了 5MB 限制，这里读进内存完全不会有 OOM 风险
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件流失败: %w", err)
	}

	targetBytes := targetSizeKB * 1024

	// 2. 设定降级梯队：从 85 画质开始，每次降级 15，最低跌到 40 止损
	qualities := []uint{85, 75, 65, 55, 40}

	var bestBuffer *bytes.Buffer

	for _, q := range qualities {
		// 每次重试前，清空旧的缓冲区，并从内存数据重新创建一个全新的 Reader
		outputBuffer := new(bytes.Buffer)
		dataReader := bytes.NewReader(fileData)

		// 调用 cwebp 在纯内存中压缩
		err := webpbin.NewCWebP().
			Quality(q).
			Input(dataReader).
			Output(outputBuffer).
			Run()

		if err != nil {
			return nil, fmt.Errorf("WebP 压缩失败 (画质 %d): %w", q, err)
		}

		bestBuffer = outputBuffer
		currentSize := int64(outputBuffer.Len())

		// 3. 体积达标测试
		if currentSize <= targetBytes {
			// 恭喜，压缩后体积已经小于目标值，直接出栈！
			// fmt.Printf("✅ 压缩达标: 画质 %d, 体积 %d KB\n", q, currentSize/1024)
			return bestBuffer, nil
		}

		// fmt.Printf("⚠️ 压缩后仍超标: 画质 %d, 体积 %d KB，准备降级...\n", q, currentSize/1024)
	}

	// 4. 兜底策略：如果降到了最低画质 (40) 依然比 200KB 大，不再继续死磕！
	// 因为再压下去图片就没法看了，直接返回最低画质的版本即可。
	// fmt.Println("🏁 已达到最低画质底线，强行放行")
	return bestBuffer, nil
}
