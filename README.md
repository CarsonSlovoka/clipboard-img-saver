# clipboard-img-saver

保存剪貼簿的圖片

此外此專案可以幫助想學習用go它配C，使用windows api的人來當作參考

## 特色

- [X] 保存剪貼簿圖片
- [ ] 提供可框選區域，並將此區域保存成圖片
- [ ] 提供簡單的著色工具
  - [ ] 線條
  - [ ] 螢光筆
  - [ ] 矩形框
- [ ] 存成webp格式

## 編譯

請安裝gcc，且確定系統路徑可以抓到

```
# 透過chocolate取得gcc.exe
choco install mingw -y

# 查看gcc路徑位置
(gcm gcc).Source
# C:\ProgramData\chocolatey\bin\gcc.exe
```

[chocolatey: mingw](https://community.chocolatey.org/packages/mingw)

---

編譯

```
go build -o clipboard_image_saver.exe main.go
```

---

請確保`CGO_ENABLED`是開啟的

```yaml
# 查看 CGO_ENABLED 設定
go env CGO_ENABLED
# 1
```

`CGO_ENABLED=1`其實是go預設的設定，因此只需要安裝好gcc即可

## 相關知識

MinGW（Minimalist GNU for Windows）和 GCC（GNU Compiler **Collection**）有密切的關係，

因為 MinGW 是一個 Windows 平台上的開發環境，它主要用於提供一組適用於 Windows 的工具，其中包含了 GCC 編譯器

---

1. GCC 編譯器的移植：MinGW 包含了 GCC 的移植版本，使得開發者可以在 Windows 系統上使用 GCC 來編譯 C、C++ 等語言的程式。這讓 Windows 使用者可以享受 GCC 編譯器的強大功能，而不需要在 Linux 或其他 UNIX 系統上進行開發
2. Windows API 支援：MinGW 包括了一些基本的 Windows API 函式庫，允許使用者在編寫 C/C++ 程式時能夠直接訪問 Windows 的系統功能，而無需其他第三方的依賴。這些 API 是標準 Windows 應用程式所需的基本函式庫
3. 編譯器與鏈接器：除了 GCC 編譯器，MinGW 還包含了其他工具，例如 ld（鏈接器）、as（彙編器），用於完成整個編譯和鏈接過程，從而在 Windows 上生成可執行檔
4. 與 MSYS 的配合：通常，MinGW 和 MSYS（Minimal SYStem）工具包一起使用。MSYS 提供了 UNIX 命令行工具（例如 bash、make 等），讓 Windows 平台上的開發體驗更接近於 UNIX 環境，使 GCC 和其他開發工具更易於使用

