//go:generate go run . --pkgdir=.. github.com/CarsonSlovoka/clipboard-img-saver/build

package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/CarsonSlovoka/clipboard-img-saver/app"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

func ZipSource() error {
	zipName := fmt.Sprintf("../bin/%s_%s_%s_%s.zip", app.Name,
		runtime.GOOS, runtime.GOARCH, app.Version)
	f, err := os.Create(zipName)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()

		// 計算zip的hash
		bytesZip, _ := os.ReadFile(zipName)
		hasher256 := sha256.New()
		hasher256.Write(bytesZip)
		hashStr := strings.ToUpper(hex.EncodeToString(hasher256.Sum(nil)))
		hashContent := fmt.Sprintf("hashes.sha256\n\n%s\n%s\n", filepath.Base(zipName), hashStr)
		err = os.WriteFile("../bin/hash.md", []byte(hashContent), 0666) // 可以寫在release上，讓大家自行查看，能確認是否因網路下載而被中途串改資料
		if err != nil {
			log.Fatal(err)
		}
		log.Println("zipSource done!")
	}()

	zipWriter := zip.NewWriter(f)
	defer func() {
		log.Println("closing zip archive...")
		_ = zipWriter.Close()
	}()

	var w io.Writer
	for _, d := range []struct {
		srcPath string
		outPath string
	}{
		{fmt.Sprintf("../bin/%s.exe", app.ExeName), app.ExeName + ".exe"},
	} {
		log.Printf("opening %q...\n", d.srcPath)
		f, err = os.Open(d.srcPath)
		if err != nil {
			return err
		}
		log.Printf("writing %q to archive...\n", d.outPath)
		w, err = zipWriter.Create(d.outPath)
		if err != nil {
			return err
		}
		if _, err = io.Copy(w, f); err != nil {
			return err
		}
		_ = f.Close()
	}
	return nil
}

func NewCmd(name, dir string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func main() {
	if err, removeAllTmplFunc := BuildAllTmpl(
		"app.manifest.gotmpl", "resources.rc.gotmpl",
		&AppInfoContext{
			cfg.Version,
			cfg.ExeName,
			cfg.Info.Desc,
			cfg.Info.ProductName,
			cfg.Info.Copyright,
			cfg.Info.Lang,
			cfg.Info.RequireAdmin}); err != nil {
		log.Fatal(err)
	} else {
		defer func() {
			if err = removeAllTmplFunc(); err != nil {
				log.Fatal(err)
			}
		}()
	}

	if err := os.MkdirAll("../bin", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	for _, cmd := range []*exec.Cmd{
		// create syso
		NewCmd("rsrc", ".", "-manifest", "app.manifest",
			"-o", "../bin/app.syso",
			// "-ico", "app.ico", // 不要用syso加入圖標，可能會沒辦法執行，建議圖標還是交給ResourceHacker來弄
		),

		// go build的時候如果存在app.syso，就會自動抓syso裡面的相關資訊(圖標, admin盾牌等等)
		NewCmd("go", "..",
			"build",
			// "-ldflags", "-s -w",
			// "-tags", "",
			"-o", "./bin/"+app.ExeName+".exe", // 最後一層目錄不存在會自己建立
			"-pkgdir", ".",
		),

		// "resource.rc" to "resource.res"
		NewCmd("ResourceHacker", ".",
			"-open", "resources.rc", "-save", "../bin/resources.res",
			"-action", "compile",
			"-log", "CONSOLE",
		),

		// AddVersionInfo
		NewCmd(
			"ResourceHacker", "../bin",
			"-open", app.ExeName+".exe", "-save", app.ExeName+".exe",
			"-resource", "resources.res",
			"-action", "addoverwrite",
			"-mask", "VersionInf",
			"-log", "CONSOLE",
		),
		// Add icon
		NewCmd(
			"ResourceHacker", "../bin",
			"-open", app.ExeName+".exe", "-save", app.ExeName+".exe",
			"-resource", "../build/app.ico",
			"-mask", "ICONGROUP,MAINICON,", // 注意icon的mask後面要有","不然會失敗
			"-action", "addoverwrite",
			"-log", "CONSOLE",
		),
	} {
		log.Println(YText(strings.Join(cmd.Args, " ")))
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}

	if err := ZipSource(); err != nil {
		log.Fatal(err)
	}
}
