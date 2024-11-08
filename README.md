# clipboard-img-saver

保存剪貼簿的圖片

此外此專案可以幫助想學習用go搭配C，使用windows api的人來當作參考

## 特色

- [X] 保存剪貼簿圖片
- [ ] 提供可框選區域，並將此區域保存成圖片
- [ ] 提供簡單的著色工具
  - [ ] 線條
  - [ ] 螢光筆
  - [ ] 矩形框
- 保存格式
  - [x] webp: 需要系統有`cwebp.exe`, 可從[libwebp-1.4.0-windows-x64.zip]取得
  - [x] bmp


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
git clone https://github.com/CarsonSlovoka/clipboard-img-saver.git
cd clipboard-img-saver
go build -o clipboard_image_saver.exe .
```

---

請確保`CGO_ENABLED`是開啟的

```yaml
# 查看 CGO_ENABLED 設定
go env CGO_ENABLED
# 1
```

`CGO_ENABLED=1`其實是go預設的設定，因此只需要安裝好gcc即可

## USAGE

```yaml
# go build -o clipboard_image_saver.exe .
# 你可以考慮換個比較簡單的執行檔名稱
# go build -o cis.exe .
cis.exe -help # 查看幫助
cis.exe -q 20 # 以webp輸出, quality 20, (預設75)
cis.exe -o "C:\myOutputDir"  # 指定輸出目錄, 預設為當前目錄
cis.exe -format .bmp  # 輸出成bmp格式
```


## webp

你可以從這些連結來了解webp

- https://github.com/webmproject/libwebp/tree/main
- https://developers.google.com/speed/webp?hl=zh-tw

[下載列表](https://storage.googleapis.com/downloads.webmproject.org/releases/webp/index.html)

[libwebp-1.4.0-windows-x64.zip]

你可以使用裡面的include, lib來鑲嵌，但是他是用MSVC(Microsoft Visual C++)去編譯，所以如果你用的是MinGW(Minimalist GNU for Windows)會編譯失敗

也可以善用裡面bin目錄提供的執行檔即可

[libwebp-1.4.0-windows-x64/bin/cwebp.exe](https://github.com/webmproject/libwebp/blob/f999d94/doc/tools.md#encoding-tool)

```yaml
cwebp input.png -q 80 -o output.webp  # q為輸出質量, 預設75範圍從[0~100]
```

## 相關知識

MinGW（Minimalist GNU for Windows）和 GCC（GNU Compiler **Collection**）有密切的關係，

因為 MinGW 是一個 Windows 平台上的開發環境，它主要用於提供一組適用於 Windows 的工具，其中包含了 GCC 編譯器

---

1. GCC 編譯器的移植：MinGW 包含了 GCC 的移植版本，使得開發者可以在 Windows 系統上使用 GCC 來編譯 C、C++ 等語言的程式。這讓 Windows 使用者可以享受 GCC 編譯器的強大功能，而不需要在 Linux 或其他 UNIX 系統上進行開發
2. Windows API 支援：MinGW 包括了一些基本的 Windows API 函式庫，允許使用者在編寫 C/C++ 程式時能夠直接訪問 Windows 的系統功能，而無需其他第三方的依賴。這些 API 是標準 Windows 應用程式所需的基本函式庫
3. 編譯器與鏈接器：除了 GCC 編譯器，MinGW 還包含了其他工具，例如 ld（鏈接器）、as（彙編器），用於完成整個編譯和鏈接過程，從而在 Windows 上生成可執行檔
4. 與 MSYS 的配合：通常，MinGW 和 MSYS（Minimal SYStem）工具包一起使用。MSYS 提供了 UNIX 命令行工具（例如 bash、make 等），讓 Windows 平台上的開發體驗更接近於 UNIX 環境，使 GCC 和其他開發工具更易於使用


[libwebp-1.4.0-windows-x64.zip]: https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-1.4.0-windows-x64.zip
