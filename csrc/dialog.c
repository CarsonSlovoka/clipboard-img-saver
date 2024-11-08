#include <windows.h>
#include <commdlg.h> // #cgo windows LDFLAGS: -lcomdlg32
#include <wchar.h>
#include <stdlib.h>


wchar_t* ShowSaveFileDialog(
    const wchar_t* initialDir,
    const wchar_t* title,
    const wchar_t* filter, const wchar_t* defExt  // 結尾不像go有多的,
) {
    OPENFILENAMEW ofn;
    ZeroMemory(&ofn, sizeof(ofn));
    wchar_t* szFile = (wchar_t*)malloc(sizeof(wchar_t) * MAX_PATH);
    if (szFile == NULL) {
        return NULL;
    }
    szFile[0] = '\0';

    ofn.lStructSize = sizeof(ofn);
    ofn.hwndOwner = NULL;
    ofn.lpstrFilter = filter;
    ofn.lpstrFile = szFile;
    ofn.nMaxFile = MAX_PATH;
    ofn.lpstrTitle = title;
    ofn.Flags = OFN_OVERWRITEPROMPT;
    ofn.lpstrDefExt = defExt;

    // 初始保存路徑
    ofn.lpstrInitialDir = initialDir;

    if (GetSaveFileNameW(&ofn)) {
        return szFile;
    } else {
        // 失敗或取消
        free(szFile);
        return NULL;
    }
}
