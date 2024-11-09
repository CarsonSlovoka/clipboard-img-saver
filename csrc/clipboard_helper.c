#include <windows.h>

// 打開剪貼簿並獲取位圖數據
HBITMAP GetClipboardBitmap() {
    HBITMAP hBitmap = NULL;
    if (OpenClipboard(NULL)) {
        // 當使用截圖按鈕`Print Screen Key`，剪貼簿裡面可能會包含其他格式，例如CF_DIB或者CF_DIBV5，僅用CF_BITMAP去檢驗會不夠完整
        // 因此最後會導致獲取到位圖失敗，或者得到NULL
        // HANDLE hBitmap = GetClipboardData(CF_BITMAP);

        if (IsClipboardFormatAvailable(CF_BITMAP)) {
            hBitmap = (HBITMAP)GetClipboardData(CF_BITMAP);
            // hBitmap = (HBITMAP)CopyImage(GetClipboardData(CF_BITMAP), IMAGE_BITMAP, 0, 0, LR_COPYRETURNORG);
        } else if (IsClipboardFormatAvailable(CF_DIB) || IsClipboardFormatAvailable(CF_DIBV5)) {
            UINT format = IsClipboardFormatAvailable(CF_DIB) ? CF_DIB : CF_DIBV5;
            printf("format: %d\n", format);
            HANDLE hDIB = GetClipboardData(format);
            if (hDIB != NULL) {
                // 將 DIB 轉為 HBITMAP
                BITMAPINFOHEADER* bih = (BITMAPINFOHEADER*)GlobalLock(hDIB);
                if (bih != NULL) {
                    HDC hdc = GetDC(NULL);
                    // hBitmap = CreateDIBitmap(hdc, bih, CBM_INIT, (BYTE*)bih + bih->biSize + bih->biClrUsed * sizeof(RGBQUAD), (BITMAPINFO*)bih, DIB_RGB_COLORS);
                    void* bits = ((BYTE*)bih) + bih->biSize + (bih->biClrUsed * sizeof(RGBQUAD));
                    BITMAPINFO* bmi = (BITMAPINFO*)bih;
                    hBitmap = CreateDIBitmap(hdc, bih, CBM_INIT, bits, bmi, DIB_RGB_COLORS);
                    ReleaseDC(NULL, hdc);
                    GlobalUnlock(hDIB);
                }
            }
        }

        if (!CloseClipboard()) {
            printf("Failed to close clipboard, error: %ld\n", GetLastError()); // <stdio.h>
        }
        return (HBITMAP)hBitmap;
    } else {
      printf("Failed to open clipboard, error: %ld\n", GetLastError());
    }
    return hBitmap;
}

// 提供一個輔助函數來獲取 HBITMAP 作為 HANDLE
HANDLE GetBitmapHandle(HBITMAP hBitmap) {
    return (HANDLE)hBitmap;
}
