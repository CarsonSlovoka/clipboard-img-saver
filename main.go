package main

/*
#cgo windows CFLAGS: -DUNICODE
// #cgo CFLAGS: -I./csrc -I./cutils  // 表示也會在csrc, cutils目錄之中尋找c文件
// #cgo windows LDFLAGS: -lgdi32 -luser32 // user32可以不需要
#cgo windows LDFLAGS: -lgdi32
#cgo CFLAGS: -I./csrc
#include "clipboard_helper.c"
*/
import "C"
import (
	"flag"
	"fmt"
	"github.com/CarsonSlovoka/clipboard-img-saver/app"
	"log"
	"os"
	"path/filepath"
	"unsafe"
)

var clipboardChanged chan bool

func init() {
	clipboardChanged = make(chan bool)
}

func EmptyClipboard() {
	if C.OpenClipboard(C.HWND(C.NULL)) != 0 {
		C.EmptyClipboard()
		C.CloseClipboard()
	}
}

func saveClipboardImage(outputDir string) error {
	hBitmap := C.GetClipboardBitmap()
	if hBitmap == nil {
		return fmt.Errorf("剪貼簿中沒有圖片數據")
	}

	// 創建位圖信息頭
	// https://learn.microsoft.com/en-us/windows/win32/api/wingdi/ns-wingdi-bitmap
	var bitmap C.BITMAP

	/*
		C.GetObject(
			C.HANDLE(hBitmap),  // cannot convert hBitmap (variable of type _Ctype_HBITMAP) to type _Ctype_HANDLE
			C.sizeof_BITMAP,
			unsafe.Pointer(&bitmap), // cannot use _cgo2 (variable of type unsafe.Pointer) as _Ctype_LPVOID value in argument to _Cfunc_GetObject
		)
	*/
	// 使用 GetBitmapHandle 來獲取 HANDLE 類型
	hBitmapHandle := C.GetBitmapHandle(hBitmap)
	C.GetObject(hBitmapHandle, C.sizeof_BITMAP, C.LPVOID(unsafe.Pointer(&bitmap)))

	pixelDataSize := int(bitmap.bmWidthBytes) * int(bitmap.bmHeight)
	bitmap.bmHeight *= -1 // 調整 BMP 頭信息中的高度：將位圖的高度設置為負值，表示數據從上到下排列

	// BMP 文件頭
	fileHeader := make([]byte, 14)
	infoHeader := make([]byte, 40)

	// 填寫 BMP 文件頭 (BITMAPFILEHEADER)
	//  all the integer values are stored in "little-endian" format
	fileHeader[0] = 'B'
	fileHeader[1] = 'M'
	fileSize := 14 + 40 + int(bitmap.bmWidthBytes)*int(bitmap.bmHeight)
	fileHeader[2] = byte(fileSize)
	fileHeader[3] = byte(fileSize >> 8)
	fileHeader[4] = byte(fileSize >> 16)
	fileHeader[5] = byte(fileSize >> 24)
	fileHeader[10] = 14 + 40

	// 填寫 BMP 信息頭 (BITMAPINFOHEADER)
	infoHeader[0] = 40 // 信息頭大小
	infoHeader[4] = byte(bitmap.bmWidth)
	infoHeader[5] = byte(bitmap.bmWidth >> 8)
	infoHeader[6] = byte(bitmap.bmWidth >> 16)
	infoHeader[7] = byte(bitmap.bmWidth >> 24)
	infoHeader[8] = byte(bitmap.bmHeight)
	infoHeader[9] = byte(bitmap.bmHeight >> 8)
	infoHeader[10] = byte(bitmap.bmHeight >> 16)
	infoHeader[11] = byte(bitmap.bmHeight >> 24)
	infoHeader[12] = 1 // 平面數
	infoHeader[14] = byte(bitmap.bmBitsPixel)

	// 使用 Windows API 彈出對話框選擇文件名
	var outputPath string
	fmt.Println("請輸入文件名（不包括路徑，將保存於指定目錄）：")
	if _, err := fmt.Scanln(&outputPath); err != nil {
		return err
	}
	if filepath.Ext(outputPath) == "" {
		outputPath += ".bmp"
	}
	outputPath = filepath.Join(outputDir, outputPath)

	// 打開文件並寫入 BMP 數據
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = file.Write(fileHeader)
	if err != nil {
		return err
	}
	_, err = file.Write(infoHeader)
	if err != nil {
		return err
	}

	// 寫入像素數據
	pixelData := make([]byte, pixelDataSize)
	// C.GetBitmapBits(C.HBITMAP(hBitmap), C.LONG(pixelDataSize), unsafe.Pointer(&pixelData[0])) // cannot use _cgo2 (variable of type unsafe.Pointer) as _Ctype_LPVOID value in argument to _Cfunc_GetBitmapBits
	C.GetBitmapBits(C.HBITMAP(hBitmap), C.LONG(pixelDataSize), C.LPVOID(unsafe.Pointer(&pixelData[0])))
	_, err = file.Write(pixelData)
	if err != nil {
		return err
	}
	absOutputPath, _ := filepath.Abs(outputPath)
	log.Printf("圖片已順利保存至:%q", absOutputPath)
	EmptyClipboard()

	return nil
}

func init() {
	print(app.About())
}

func main() {
	var outputDir string
	flag.StringVar(&outputDir, "output", ".", "指定圖片保存的目錄")
	flag.Parse()
	if outputDir == "" {
		fmt.Println("必須指定輸出目錄，使用 -output 參數來指定")
		return
	}
	fmt.Printf("圖片輸出目錄: %s\n", outputDir)
	err := saveClipboardImage(outputDir)
	if err != nil {
		log.Printf("錯誤: %s\n", err)
	}
}
