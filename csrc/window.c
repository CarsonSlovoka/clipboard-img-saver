// WNDPROC myDefWindowProcW = DefWindowProcW;

const UINT WM_IMG_SAVE = WM_USER + 1;

// 定義一個簡單的窗口過程
LRESULT CALLBACK MyWindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam) {
    if (uMsg == WM_CLIPBOARDUPDATE) {
        PostMessage(hwnd, WM_IMG_SAVE, 0, 0);
    }
    return DefWindowProc(hwnd, uMsg, wParam, lParam);
}

WNDPROC GetMyWindowProc() {
    return MyWindowProc;
}
