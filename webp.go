package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os/exec"
)

// convertToWebP 使用 cwebp 將 PNG 圖像數據轉為 WebP 格式
func convertToWebP(img any, quality uint8) ([]byte, error) {
	cmd := exec.Command("cwebp", "-q", fmt.Sprint(quality),
		// https://developers.google.com/speed/webp/docs/dwebp
		"-o", "-", // -o - 表示將輸出寫入stdout
		"--", "-", // - 表示輸入是從stdin
	)
	var output bytes.Buffer
	cmd.Stdout = &output

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	// 寫入 stdin
	go func() {
		switch v := img.(type) {
		case image.Image:
			_ = png.Encode(stdin, v)
		case []byte:
			_, _ = stdin.Write(v)
		}
		_ = stdin.Close()
	}()

	if err = cmd.Wait(); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}
