#include <windows.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

typedef void* LPVOID;

// 打開剪貼簿並獲取位圖數據
HBITMAP GetClipboardBitmap() {
    if (OpenClipboard(NULL)) {
        HANDLE hBitmap = GetClipboardData(CF_BITMAP);
        CloseClipboard();
        return (HBITMAP)hBitmap;
    }
    return NULL;
}

// 提供一個輔助函數來獲取 HBITMAP 作為 HANDLE
HANDLE GetBitmapHandle(HBITMAP hBitmap) {
    return (HANDLE)hBitmap;
}
