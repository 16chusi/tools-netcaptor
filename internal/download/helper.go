package download

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// DownloadFile 下载文件并弹出保存对话框
func DownloadFile(ctx context.Context, url string, defaultFilename string) error {
	// 弹出保存对话框
	savePath, err := runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Title:           "保存文件",
	})

	if err != nil {
		return err
	}

	if savePath == "" {
		return nil // 用户取消
	}

	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 创建文件
	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 写入文件
	_, err = io.Copy(out, resp.Body)
	return err
}
