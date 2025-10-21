package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenInChrome(url string, proxyURL string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "chrome",
			fmt.Sprintf("--proxy-server=%s", proxyURL),
			"--new-window",
			url)
	case "darwin":
		cmd = exec.Command("open", "-a", "Google Chrome",
			"--args",
			fmt.Sprintf("--proxy-server=%s", proxyURL),
			url)
	case "linux":
		cmd = exec.Command("google-chrome",
			fmt.Sprintf("--proxy-server=%s", proxyURL),
			"--ignore-certificate-errors",
			"--new-window",
			url)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return cmd.Start()
}

func OpenInEdge(url string, proxyURL string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "msedge",
			fmt.Sprintf("--proxy-server=%s", proxyURL),
			"--new-window",
			url)
	case "darwin":
		cmd = exec.Command("open", "-a", "Microsoft Edge",
			"--args",
			fmt.Sprintf("--proxy-server=%s", proxyURL),
			url)
	case "linux":
		cmd = exec.Command("microsoft-edge",
			fmt.Sprintf("--proxy-server=%s", proxyURL),
			"--new-window",
			url)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return cmd.Start()
}

func OpenInFirefox(url string, proxyURL string) error {
	// Firefox 代理配置较复杂，暂时使用默认方式打开
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "firefox", url)
	case "darwin":
		cmd = exec.Command("open", "-a", "Firefox", url)
	case "linux":
		cmd = exec.Command("firefox", url)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return cmd.Start()
}
