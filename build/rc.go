package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CarsonSlovoka/clipboard-img-saver/app"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	textTemplate "text/template"
)

type Config struct {
	Version string
	ExeName string
	LdFlags string
	Info    struct { // 此為用於填充詳細資料所用
		Desc         string
		ProductName  string
		RequireAdmin bool
		Copyright    string
		Lang         string
	}
}

var cfg Config

func init() {
	f, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = f.Close()
	}()
	decoder := json.NewDecoder(f)
	if err = decoder.Decode(&cfg); err != nil {
		panic(err)
	}
	cfg.ExeName = strings.ReplaceAll(app.ExeName, " ", "-") // rc檔案裡面用到的名稱不能有空白
	cfg.Version = strings.ReplaceAll(app.Version, "v", "")
}

// TextColor 可以更改console的顏色
func TextColor(fr, fg, fb, br, bg, bb int) func(msg string) string {
	return func(msg string) string {
		return fmt.Sprintf("\u001B[48;2;%d;%d;%dm\u001B[38;2;%d;%d;%dm%s\u001B[0m",
			br, bg, bb,
			fr, fg, fb,
			msg,
		)
	}
}

var (
	YText func(msg string) string
)

func init() {
	YText = TextColor(0, 0, 0, 255, 255, 0)
}

var funcMap = map[string]any{
	"ternary": func(condition bool, trueVal, falseVal any) any {
		if condition {
			return trueVal
		}
		return falseVal
	},
	"replaceAll": func(s, old, new string) string {
		return strings.ReplaceAll(s, old, new)
	},
	"dict": func(values ...any) (map[string]any, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("parameters must be even")
		}
		dict := make(map[string]any)
		var key, val any
		for {
			key, val, values = values[0], values[1], values[2:]
			switch reflect.ValueOf(key).Kind() {
			case reflect.String:
				dict[key.(string)] = val
			default:
				return nil, errors.New(`type must equal to "string"`)
			}
			if len(values) == 0 {
				break
			}
		}
		return dict, nil
	},
	// 自動填補到4碼的版號
	"makeValidVersion": func(version string) string {
		validVersion := version
		for i := 0; i < 3-strings.Count(version, "."); i++ {
			validVersion += ".0"
		}
		return validVersion
	},
}

type AppInfoContext struct {
	Version      string
	ExeName      string
	Desc         string
	ProductName  string
	Copyright    string
	Translation  string
	RequireAdmin bool
}

// BuildAllTmpl 提供app.manifest, resources.rc等樣版路徑，建立出相對應的檔案
// 如果要刪除所產生出來的產物，可以呼叫deferFunc
func BuildAllTmpl(manifestPath, resourcesPath string, appInfoCtx *AppInfoContext) (err error, deferFunc func() error) {
	if !strings.HasSuffix(manifestPath, ".gotmpl") || !strings.HasSuffix(resourcesPath, ".gotmpl") {
		return fmt.Errorf("please make sure the file suffix is %q", ".gotmpl"), nil
	}

	tmplPaths := make([]string, 0) // 紀錄產生出來的tmp檔案，最後要移除
	for _, tmplPath := range []string{manifestPath, resourcesPath} {
		outFilePath := tmplPath[:len(tmplPath)-7] // app.manifest.gotmpl => app.manifest
		var outF *os.File
		outF, err = os.OpenFile(outFilePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return err, nil
		}
		tmplPaths = append(tmplPaths, outFilePath)
		t := textTemplate.Must(
			textTemplate.New(filepath.Base(tmplPath)).Funcs(funcMap).ParseFiles(tmplPath),
		)
		if err = t.Execute(outF, appInfoCtx); err != nil {
			return err, nil
		}
		if err = outF.Close(); err != nil {
			return err, nil
		}
	}
	return nil, func() error {
		for _, curF := range tmplPaths {
			if err = os.Remove(curF); err != nil {
				return err
			}
		}
		return nil
	}
}
