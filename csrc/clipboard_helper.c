#include <stdio.h>
#include <windows.h>

// 檢查剪貼簿內容是否為圖片
BOOL IsClipboardImage() {
    if (!OpenClipboard(NULL)) {
        return FALSE;
    }
    BOOL isImage = IsClipboardFormatAvailable(CF_BITMAP) || IsClipboardFormatAvailable(CF_DIB) || IsClipboardFormatAvailable(CF_DIBV5);
    CloseClipboard();
    return isImage;
}

typedef struct {
    HGLOBAL data;
    DWORD size;
} BitmapMemory;


HBITMAP HandleToHBitmap(HANDLE h) { // 解決: cannot convert hBitmap (variable of type _Ctype_HANDLE) to type _Ctype_HBITMAP
    return (HBITMAP)h;
}

BitmapMemory SaveImageToMemory(HBITMAP hBitmap) {
    BitmapMemory result;
    result.data = NULL;
    result.size = 0;

    BITMAP bmp;
    // https://learn.microsoft.com/zh-tw/windows/win32/api/wingdi/ns-wingdi-bitmapinfo
    PBITMAPINFO pbmi;
    WORD cClrBits;
    PBITMAPINFOHEADER pbih;
    LPBYTE lpBits;
    HDC hDC;

    hDC = CreateCompatibleDC(NULL);
    GetObject(hBitmap, sizeof(BITMAP), (LPSTR)&bmp);

    cClrBits = (WORD)(bmp.bmPlanes * bmp.bmBitsPixel);
    if (cClrBits == 1)
        cClrBits = 1;
    else if (cClrBits <= 4)
        cClrBits = 4;
    else if (cClrBits <= 8)
        cClrBits = 8;
    else if (cClrBits <= 16)
        cClrBits = 16;
    else if (cClrBits <= 24)
        cClrBits = 24;
    else
        cClrBits = 32;

    if (cClrBits != 24)
        pbmi = (PBITMAPINFO)LocalAlloc(LPTR, sizeof(BITMAPINFOHEADER) + sizeof(RGBQUAD) * (1 << cClrBits));
    else
        pbmi = (PBITMAPINFO)LocalAlloc(LPTR, sizeof(BITMAPINFOHEADER));

    pbih = (PBITMAPINFOHEADER)pbmi;
    pbih->biSize = sizeof(BITMAPINFOHEADER);
    pbih->biWidth = bmp.bmWidth;
    pbih->biHeight = bmp.bmHeight;
    pbih->biPlanes = bmp.bmPlanes;
    pbih->biBitCount = bmp.bmBitsPixel;
    if (cClrBits < 24)
        pbih->biClrUsed = (1 << cClrBits);
    pbih->biCompression = BI_RGB;
    // https://learn.microsoft.com/en-us/windows/win32/api/wingdi/ns-wingdi-bitmapinfoheader#calculating-surface-stride
    // ~在c語言是取反的意思(0變1, 1變0)
    // 4byte對齊。 (x+31)&^31 不足32bit就多給1個4byte; &~31可以去除尾數; >> 3 相當於除上8 也就是單位從bit換成byte
    pbih->biSizeImage = (((pbih->biWidth * cClrBits + 31) & ~31) >> 3) * pbih->biHeight;
    pbih->biClrImportant = 0;

    lpBits = (LPBYTE)GlobalAlloc(GMEM_FIXED, pbih->biSizeImage);
    if (GetDIBits(hDC, hBitmap, 0, (WORD)pbih->biHeight, lpBits, pbmi, DIB_RGB_COLORS)) {
        DWORD totalSize = sizeof(BITMAPFILEHEADER) + pbih->biSize + pbih->biClrUsed * sizeof(RGBQUAD) + pbih->biSizeImage;

        // 分配緩衝區並將數據拷貝進去
        result.data = (LPBYTE)GlobalAlloc(GMEM_FIXED, totalSize);
        if (result.data != NULL) {
            // https://learn.microsoft.com/zh-tw/windows/win32/api/wingdi/ns-wingdi-bitmapfileheader
            BITMAPFILEHEADER hdr;
            hdr.bfType = 0x4d42; // "BM"
            hdr.bfSize = totalSize;
            hdr.bfReserved1 = 0;
            hdr.bfReserved2 = 0;
            hdr.bfOffBits = (DWORD)sizeof(BITMAPFILEHEADER) + pbih->biSize + pbih->biClrUsed * sizeof(RGBQUAD);

            // 文件頭、位圖信息頭和位圖數據
            LPBYTE p = result.data;
            memcpy(p, &hdr, sizeof(BITMAPFILEHEADER));
            p += sizeof(BITMAPFILEHEADER);
            memcpy(p, pbih, sizeof(BITMAPINFOHEADER) + pbih->biClrUsed * sizeof(RGBQUAD));
            p += sizeof(BITMAPINFOHEADER) + pbih->biClrUsed * sizeof(RGBQUAD);
            memcpy(p, lpBits, pbih->biSizeImage);

            result.size = totalSize;
        }
    }
    GlobalFree((HGLOBAL)lpBits);
    LocalFree(pbmi);
    DeleteDC(hDC);

    return result;
}
