package main

/*
#cgo windows CFLAGS: -DUNICODE
// #cgo CFLAGS: -I./csrc -I./cutils  // 表示也會在csrc, cutils目錄之中尋找c文件
// #cgo windows LDFLAGS: -lgdi32 -luser32 // user32可以不需要
#cgo windows LDFLAGS: -lgdi32 -lcomdlg32
#cgo CFLAGS: -I./csrc
#include "clipboard_helper.c"
#include "window.c"
#include "dialog.c"
*/
import "C"
import (
	"bytes"
	"flag"
	"fmt"
	"github.com/CarsonSlovoka/clipboard-img-saver/app"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

const MAX_PATH = 260
const ERROR_CANCELLED = 1223

func EmptyClipboard() {
	if C.OpenClipboard(C.HWND(C.NULL)) != 0 {
		C.EmptyClipboard()
		C.CloseClipboard()
	}
}

func saveClipboardImage(outputDir, format string, quality uint8) error {
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

	// outputPath, err := getSaveFileNameByScan()
	outputPath, err := GetSaveFileName("save image", outputDir, format)
	if err != nil {
		if err.(syscall.Errno) == ERROR_CANCELLED {
			log.Printf("取消存檔")
			return nil
		}
		return err
	}

	buf := bytes.NewBuffer(nil)

	_, err = buf.Write(fileHeader)
	if err != nil {
		return err
	}
	_, err = buf.Write(infoHeader)
	if err != nil {
		return err
	}

	// 寫入像素數據
	pixelData := make([]byte, pixelDataSize)
	// C.GetBitmapBits(C.HBITMAP(hBitmap), C.LONG(pixelDataSize), unsafe.Pointer(&pixelData[0])) // cannot use _cgo2 (variable of type unsafe.Pointer) as _Ctype_LPVOID value in argument to _Cfunc_GetBitmapBits
	C.GetBitmapBits(C.HBITMAP(hBitmap), C.LONG(pixelDataSize), C.LPVOID(unsafe.Pointer(&pixelData[0])))
	_, err = buf.Write(pixelData)
	if err != nil {
		return err
	}

	var result []byte
	if format == ".webp" {
		if filepath.Ext(outputPath) == "" {
			outputPath += ".webp"
		}
		result, err = convertToWebP(buf.Bytes(), quality)
		if err != nil {
			return err
		}
	} else {
		// 一律存成bmp
		if filepath.Ext(outputPath) == "" {
			outputPath += ".bmp"
		}
		result = buf.Bytes()
	}

	// outputPath = filepath.Join(outputDir, outputPath)  // 如果使用GetSaveFileNameW，出來的路徑已經是絕對路徑
	var file *os.File
	file, err = os.Create(outputPath)
	if err != nil {
		return err
	}
	_, _ = file.Write(result)
	_ = file.Close()

	absOutputPath, _ := filepath.Abs(outputPath)
	log.Printf("圖片已順利保存至:%q", absOutputPath)
	// EmptyClipboard()

	return nil
}

func getSaveFileNameByScan() (string, error) {
	var outputPath string
	fmt.Println("請輸入文件名（不包括路徑，將保存於指定目錄）：")
	if _, err := fmt.Scanln(&outputPath); err != nil {
		/*
			if err.Error() == "unexpected newline" {
				log.Printf("檔名為空，跳過不存檔")
				return "", nil
			}
		*/
		return "", err
	}
	return outputPath, nil
}

func GetSaveFileName(title, defaultSaveDir, ext string) (string, error) {
	// image (*.webp)\0*.webp\0
	cFilter := goStrToCWideString(
		fmt.Sprintf("image (*%s)\x00*%s\x00",
			ext, ext,
		) + "All Files (*.*)\x00*.*\x00\x00",
	)
	defer C.free(unsafe.Pointer(cFilter))
	cDefExt := goStrToCWideString(ext[1:])
	defer C.free(unsafe.Pointer(cDefExt))
	cTitle := goStrToCWideString(title)
	defer C.free(unsafe.Pointer(cTitle))
	cInitialDir := goStrToCWideString(defaultSaveDir)
	defer C.free(unsafe.Pointer(cInitialDir))

	cFileName := C.ShowSaveFileDialog(cInitialDir, cTitle, cFilter, cDefExt)
	if cFileName == nil {
		errCode := C.CommDlgExtendedError()
		if errCode == 0 {
			return "", syscall.Errno(ERROR_CANCELLED)
		}
		return "", fmt.Errorf("GetSaveFileNameW failed with error code %w", syscall.Errno(errCode))
	}
	defer C.free(unsafe.Pointer(cFileName))
	outputPath := wcharPtrToString(cFileName)
	return outputPath, nil
}

var clipboardChanged chan bool

func init() {
	print(app.About())
}

func AddClipboardFormatListener() error {
	// 創建一個隱藏的窗口

	// 轉換 Go 字符串到 C 字符串
	className, _ := syscall.UTF16PtrFromString("clipboard-img-saver")
	cClassName := C.CString(C.GoString((*C.char)(unsafe.Pointer(className))))
	defer C.free(unsafe.Pointer(cClassName))

	hInstance := C.GetModuleHandle(nil)

	var wc C.WNDCLASSW
	// wc.lpfnWndProc = C.DefWindowProcW  // cannot use (_Cgo_ptr(_Cfpvar_fp_DefWindowProcW)) (value of type unsafe.Pointer) as _Ctype_WNDPROC value in assignment
	// wc.lpfnWndProc = C.myDefWindowProcW // 使用聲明的 myDefWindowProcW
	// wc.lpfnWndProc = C.MyWindowProc // cannot use (_Cgo_ptr(_Cfpvar_fp_MyWindowProc)) (value of type unsafe.Pointer) as _Ctype_WNDPROC value in assignment
	wc.lpfnWndProc = C.GetMyWindowProc()
	wc.cbClsExtra = 0
	wc.cbWndExtra = 0
	wc.hInstance = hInstance
	wc.hIcon = nil
	wc.hIcon = nil
	wc.hCursor = nil
	wc.hbrBackground = nil
	wc.lpszMenuName = nil
	// wc.lpszClassName = (*C.wchar_t)(unsafe.Pointer(className))  // runtime error: cgo argument has Go pointer to unpinned Go pointer  Go 指針不應該直接傳遞給 C 語言函數（在這種情況下是 C.RegisterClassW），因為 Go 垃圾回收器（GC）可能會移動這些指針。這在 cgo 中是一個常見問題，當 Go 指針包含 Go 分配的記憶體時，不能直接將其傳遞給 C 函數. 若 className 是單純的字符串，可以將其轉換成 C 字符串並在用完後釋放，這樣可以避免 Go 指針被傳遞給 C 函數
	wc.lpszClassName = (*C.wchar_t)(unsafe.Pointer(cClassName))

	if C.RegisterClassW(&wc) == 0 {
		return fmt.Errorf("failed to register window class")
	}

	hwnd := C.CreateWindowExW(
		0,
		(*C.wchar_t)(unsafe.Pointer(cClassName)), // 使用 UTF16 字串指標
		nil,
		0,
		0, 0, 0, 0,
		nil,
		nil,
		nil,
		nil,
	)
	if hwnd == nil {
		return fmt.Errorf("failed to create window")
	}

	success := C.AddClipboardFormatListener(hwnd)
	if success == 0 {
		return fmt.Errorf("無法添加剪貼簿監聽")
	}
	return nil
}

func main() {
	runtime.LockOSThread()

	var outputDir string
	var format string
	var quality uint
	flag.StringVar(&outputDir, "o", ".", "指定圖片保存的目錄")
	flag.StringVar(&format, "format", ".webp", "輸出的格式, .bmp, .webp")
	flag.UintVar(&quality, "q", 75, "輸出質量(僅限webp有用)")
	flag.Parse()
	if outputDir == "" {
		fmt.Println("必須指定輸出目錄，使用 -output 參數來指定")
		return
	}
	fmt.Printf("圖片輸出目錄: %s\n", outputDir)

	clipboardChanged = make(chan bool)

	// 註冊剪貼簿監聽
	go func() {
		for {
			select {
			case <-clipboardChanged:
				runtime.LockOSThread()
				err := saveClipboardImage(outputDir, format, uint8(quality))
				if err != nil {
					log.Printf("錯誤: %s\n", err) // 如果有錯誤可能就會卡住，收不到下一個GetMessage的消息. 可能是stdout有衝突?
				}
				runtime.UnlockOSThread()
			}
		}
	}()

	// 不能拿其他的窗口，要自己建立，否則GetMessage還是拿不到東西
	// windowName, _ := syscall.UTF16PtrFromString("MSPaintApp")
	// paintWindowHWND := C.FindWindowW((*C.wchar_t)(unsafe.Pointer(windowName)), (*C.wchar_t)(nil))
	// C.AddClipboardFormatListener(C.HWND(C.NULL))
	// success := C.AddClipboardFormatListener(paintWindowHWND)
	// if success == 0 {
	// 	fmt.Println("無法添加剪貼簿監聽於MSPaintApp")
	// 	return
	// }

	err := AddClipboardFormatListener()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("成功設置剪貼簿監聽")

	var msg C.MSG
	for {
		ret := C.GetMessage(&msg, nil, 0, 0) // 使用GetMessage之前一定要先創建一個窗口，不然執行到這邊就會被卡住
		// ret := C.GetMessage(&msg, paintWindowHWND, 0, 0)
		if ret <= 0 {
			break
		}
		// if msg.message == C.WM_CLIPBOARDUPDATE {
		if msg.message == C.WM_IMG_SAVE {
			clipboardChanged <- true
			log.Println("檢測到剪貼簿變更")
			// continue // 與這個無關，假設有錯誤加了這個還是會卡住
		}
		C.TranslateMessage(&msg)
		C.DispatchMessage(&msg)
	}
}
