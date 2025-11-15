package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"netcaptor/internal/browser"
	"netcaptor/internal/download"
	"netcaptor/internal/utils"

	"github.com/playwright-community/playwright-go"
)

// App struct
type App struct {
	ctx        context.Context
	browser    *browser.BrowserManager
	chromedp   *browser.ChromeDPManager
	downloader *download.Downloader
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		browser:    browser.NewBrowserManager(),
		chromedp:   browser.NewChromeDPManager(),
		downloader: download.NewDownloader("./downloads"),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// 不自动初始化浏览器，等待用户手动启动
}

// StartScraping 开始抓取指定网站的下载链接
func (a *App) StartScraping(url, linkSelector, nextPageSelector string, maxPages int) ([]string, error) {
	// 初始化浏览器
	err := a.browser.Init()
	if err != nil {
		return nil, fmt.Errorf("初始化浏览器失败: %v", err)
	}

	// 清空之前的任务
	a.downloader.ClearTasks()

	var allLinks []string
	pageCount := 0

	// 访问第一页
	err = a.browser.NavigateToURL(url)
	if err != nil {
		return nil, fmt.Errorf("访问网站失败: %v", err)
	}

	for {
		pageCount++
		utils.AppLog.Info(fmt.Sprintf("正在抓取第 %d 页...", pageCount))

		// 提取当前页面的下载链接
		links, err := a.browser.ExtractDownloadLinks(linkSelector)
		if err != nil {
			utils.AppLog.Info(fmt.Sprintf("提取链接失败: %v", err))
			break
		}

		allLinks = append(allLinks, links...)

		// 将链接添加到下载任务
		for _, link := range links {
			a.downloader.AddTask(link)
		}

		// 检查是否达到最大页数
		if maxPages > 0 && pageCount >= maxPages {
			break
		}

		// 尝试翻到下一页
		hasNext, err := a.browser.NextPage(nextPageSelector)
		if err != nil {
			utils.AppLog.Info(fmt.Sprintf("翻页失败: %v", err))
			break
		}
		if !hasNext {
			utils.AppLog.Info("没有更多页面")
			break
		}
	}

	utils.AppLog.Info(fmt.Sprintf("抓取完成，共找到 %d 个下载链接", len(allLinks)))
	return allLinks, nil
}

// StartDownload 开始下载所有任务
func (a *App) StartDownload() error {
	return a.downloader.DownloadAll()
}

// GetDownloadTasks 获取下载任务列表
func (a *App) GetDownloadTasks() []download.DownloadTask {
	return a.downloader.GetTasks()
}

// SetDownloadPath 设置下载路径
func (a *App) SetDownloadPath(path string) error {
	// 确保路径存在
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("无效的路径: %v", err)
	}

	err = os.MkdirAll(absPath, 0755)
	if err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	a.downloader = download.NewDownloader(absPath)
	return nil
}

// GetCurrentURL 获取当前浏览器页面URL
func (a *App) GetCurrentURL() (string, error) {
	// 	if a.browser == nil || a.browser.page == nil {
	// 		return "", fmt.Errorf("浏览器未初始化")
	// 	}
	// 	return a.browser.page.URL(), nil
	return "", nil
}

// PreviewURL 预览网站（打开浏览器但不抓取）
func (a *App) PreviewURL(url string) error {
	// 初始化浏览器
	err := a.browser.Init()
	if err != nil {
		return fmt.Errorf("初始化浏览器失败: %v", err)
	}

	// 访问网站
	err = a.browser.NavigateToURL(url)
	if err != nil {
		return fmt.Errorf("访问网站失败: %v", err)
	}

	return nil
}

// OpenDebugBrowser 打开调试浏览器（不注入反检测脚本）
func (a *App) OpenDebugBrowser(url string) error {
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("启动 playwright 失败: %v", err)
	}
	defer pw.Stop()

	// 创建强化的调试浏览器
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		Devtools: playwright.Bool(true),
		Args: []string{
			"--no-sandbox",
			"--disable-web-security",
			"--disable-features=VizDisplayCompositor",
			"--force-devtools-available",
			"--disable-blink-features=AutomationControlled",
			"--disable-background-timer-throttling",
			"--disable-backgrounding-occluded-windows",
			"--disable-renderer-backgrounding",
			"--disable-ipc-flooding-protection",
		},
	})
	if err != nil {
		return fmt.Errorf("启动浏览器失败: %v", err)
	}

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	})
	if err != nil {
		return fmt.Errorf("创建上下文失败: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("创建页面失败: %v", err)
	}

	// 注入对抗 disable-devtool 的脚本
	err = page.AddInitScript(playwright.Script{Content: playwright.String(`
		// 对抗 disable-devtool 的强化脚本
		Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
		Object.defineProperty(window, 'outerHeight', { get: () => window.innerHeight });
		Object.defineProperty(window, 'outerWidth', { get: () => window.innerWidth });
		
		// 设置 tkName 跳过检测
		window.ddtk = 'skip';
		window.DDTK = 'skip';
		localStorage.setItem('ddtk', 'skip');
		sessionStorage.setItem('ddtk', 'skip');
		
		// 拦截 localStorage
		const originalGetItem = Storage.prototype.getItem;
		Storage.prototype.getItem = function(key) {
			if (key === 'ddtk' || key === 'DDTK') return 'skip';
			return originalGetItem.call(this, key);
		};
		
		// 覆盖 DisableDevtool
		window.DisableDevtool = function() { return { isRunning: false }; };
		window.DisableDevtool.isRunning = false;
		
		// 完全阻止页面跳转
		window.close = function() { console.log('阻止关闭'); return false; };
		window.location.replace = function() { console.log('阻止 replace'); return false; };
		window.location.assign = function() { console.log('阻止 assign'); return false; };
		window.location.reload = function() { console.log('阻止刷新'); return false; };
		
		// 拦截 href 设置
		Object.defineProperty(window.location, 'href', {
			set: function(url) { console.log('阻止跳转到:', url); return false; },
			get: function() { return window.location.toString(); }
		});
		
		// 拦截 history API
		history.pushState = function() { console.log('阻止 pushState'); return false; };
		history.replaceState = function() { console.log('阻止 replaceState'); return false; };
		
		// 清除 disable-devtool
		setTimeout(() => {
			const scripts = document.querySelectorAll('script');
			scripts.forEach(script => {
				if (script.src && (script.src.includes('disable-devtool') || script.src.includes('theajack'))) {
					console.log('移除脚本:', script.src);
					script.remove();
				}
			});
			for (let i = 1; i < 99999; i++) {
				window.clearInterval(i);
				window.clearTimeout(i);
			}
		}, 10);
		
		// 阻止跳转事件
		window.addEventListener('beforeunload', e => { e.preventDefault(); e.stopImmediatePropagation(); return false; }, true);
		document.addEventListener('click', e => {
			if (e.target.tagName === 'A' && e.target.href && e.target.href.includes('404.html')) {
				e.preventDefault(); e.stopPropagation(); return false;
			}
		}, true);
	`)})
	if err != nil {
		return fmt.Errorf("注入脚本失败: %v", err)
	}

	_, err = page.Goto(url)
	if err != nil {
		return fmt.Errorf("访问页面失败: %v", err)
	}

	return nil
}

// OpenChromeDPBrowser 使用 ChromeDP 打开浏览器
func (a *App) OpenChromeDPBrowser(url string) error {
	err := a.chromedp.Init()
	if err != nil {
		return fmt.Errorf("初始化 ChromeDP 失败: %v", err)
	}

	err = a.chromedp.NavigateToURL(url)
	if err != nil {
		return fmt.Errorf("访问网站失败: %v", err)
	}

	return nil
}

// StartScrapingWithChromeDP 使用 ChromeDP 开始抓取
func (a *App) StartScrapingWithChromeDP(url, linkSelector, nextPageSelector string, maxPages int) ([]string, error) {
	err := a.chromedp.Init()
	if err != nil {
		return nil, fmt.Errorf("初始化 ChromeDP 失败: %v", err)
	}

	// 清空之前的任务
	a.downloader.ClearTasks()

	var allLinks []string
	pageCount := 0

	// 访问第一页
	err = a.chromedp.NavigateToURL(url)
	if err != nil {
		return nil, fmt.Errorf("访问网站失败: %v", err)
	}

	for {
		pageCount++
		utils.AppLog.Info(fmt.Sprintf("正在抓取第 %d 页...", pageCount))

		// 提取当前页面的下载链接
		links, err := a.chromedp.ExtractDownloadLinks(linkSelector)
		if err != nil {
			utils.AppLog.Info(fmt.Sprintf("提取链接失败: %v", err))
			break
		}

		allLinks = append(allLinks, links...)

		// 将链接添加到下载任务
		for _, link := range links {
			a.downloader.AddTask(link)
		}

		// 检查是否达到最大页数
		if maxPages > 0 && pageCount >= maxPages {
			break
		}

		// 尝试翻到下一页
		hasNext, err := a.chromedp.NextPage(nextPageSelector)
		if err != nil {
			utils.AppLog.Info(fmt.Sprintf("翻页失败: %v", err))
			break
		}
		if !hasNext {
			utils.AppLog.Info("没有更多页面")
			break
		}
	}

	utils.AppLog.Info(fmt.Sprintf("抓取完成，共找到 %d 个下载链接", len(allLinks)))
	return allLinks, nil
}

// CloseBrowser 关闭浏览器
func (a *App) CloseBrowser() error {
	if a.browser != nil {
		a.browser.Close()
	}
	if a.chromedp != nil {
		a.chromedp.Close()
	}
	return nil
}

// GetTestServerPort 获取测试服务器端口（已废弃，使用 NetworkApp.GetWebhookPort）
func (a *App) GetTestServerPort() int {
	return 0
}
