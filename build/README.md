## build.go功能

- 發佈到[release頁面](https://github.com/CarsonSlovoka/clipboard-img-saver/releases)
- 添加執行檔 右鍵 > 內容 > 詳細資料 資訊
- 添加圖標

## webp轉換成ico

可以使用ImageMagick來轉換

```
convert input.webp -define icon:auto-resize=64,48,32,16 output.ico
```

如果你想要透過docker來安裝ImageMagick可以考慮此[倉庫](https://github.com/dooman87/imagemagick-docker)

```
docker run -v .:/wkDir dpokidov/imagemagick /wkDir/app.webp -define icon:auto-resize=64,48,32,16 /wkDir/app.ico
```
