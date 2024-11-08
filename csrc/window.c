// WNDPROC myDefWindowProcW = DefWindowProcW;

const UINT WM_IMG_SAVE = WM_USER + 1;

// 定義一個簡單的窗口過程
LRESULT CALLBACK MyWindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam) {
    if (uMsg == WM_CLIPBOARDUPDATE) {
        // 打開剪貼簿並檢查是否包含圖片格式
        if (OpenClipboard(NULL)) {
            UINT format = EnumClipboardFormats(0);
            BOOL hasImage = FALSE;

            // 遍歷剪貼簿格式，檢查是否包含圖片格式
            while (format != 0) {
                if (format == CF_BITMAP || format == CF_DIB || format == CF_DIBV5) {
                    hasImage = TRUE;
                    break;
                }
                format = EnumClipboardFormats(format);
            }
            CloseClipboard();

            if (hasImage) {
                PostMessage(hwnd, WM_IMG_SAVE, 0, 0);
            }
        }
    }
    return DefWindowProc(hwnd, uMsg, wParam, lParam);
}

WNDPROC GetMyWindowProc() {
    return MyWindowProc;
}
