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
	if C.IsClipboardImage() != C.TRUE {
		// 剪貼簿的內容非圖片
		return nil
	}

	// 先問使用者要不要存檔，如果不保存底下的事都不用做
	// outputPath, err := getSaveFileNameByScan()
	outputPath, err := GetSaveFileName("save image", outputDir, format)
	if err != nil {
		if err.(syscall.Errno) == ERROR_CANCELLED {
			log.Printf("取消存檔")
			return nil
		}
		return err
	}

	if C.OpenClipboard(nil) == C.FALSE {
		return fmt.Errorf("剪貼簿開啟失敗")
	}
	defer C.CloseClipboard()

	handle := C.GetClipboardData(C.CF_BITMAP)
	if handle == nil {
		return fmt.Errorf("剪貼簿中沒有圖片數據")
	}
	defer C.DeleteObject(C.HGDIOBJ(handle))

	var imgBytes []byte
	bmpMemory := C.SaveImageToMemory(C.HandleToHBitmap(handle))
	if bmpMemory.data == nil {
		return fmt.Errorf("SaveImageToMemory 得到 nil")
	}
	size := int(bmpMemory.size)
	imgBytes = C.GoBytes(unsafe.Pointer(bmpMemory.data), C.int(size))
	C.GlobalFree(C.HGLOBAL(bmpMemory.data)) // 最後不用時記得把它釋放，釋放的目標就是你(LPBYTE)GlobalAlloc(GMEM_FIXED, totalSize);時候得到的產物，也就是result.data

	var result []byte
	if format == ".webp" {
		if filepath.Ext(outputPath) == "" {
			outputPath += ".webp"
		}
		result, err = convertToWebP(imgBytes, quality)
		if err != nil {
			return err
		}
	} else {
		// 一律存成bmp
		if filepath.Ext(outputPath) == "" {
			outputPath += ".bmp"
		}
		result = imgBytes
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
	filterUTF16, err := buildFilter(ext)
	if err != nil {
		return "", err
	}
	cFilter := utf16ToCWideString(filterUTF16)
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
