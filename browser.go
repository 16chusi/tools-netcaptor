package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

type BrowserManager struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	page    playwright.Page
}

func NewBrowserManager() *BrowserManager {
	return &BrowserManager{}
}

func (bm *BrowserManager) Init() error {
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("启动 playwright 失败: %v", err)
	}
	bm.pw = pw

	// 创建反检测浏览器实例
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false), // 有头浏览器
		Args: []string{
			"--no-sandbox",
			"--disable-blink-features=AutomationControlled",
			"--disable-web-security",
			"--disable-features=VizDisplayCompositor",
			"--disable-dev-shm-usage",
			"--no-first-run",
			"--disable-extensions",
			"--disable-plugins",
			"--disable-images",
			"--disable-javascript-harmony-shipping",
			"--disable-background-timer-throttling",
			"--disable-backgrounding-occluded-windows",
			"--disable-renderer-backgrounding",
			"--disable-field-trial-config",
			"--disable-back-forward-cache",
			"--disable-ipc-flooding-protection",
			"--force-devtools-available", // 强制启用开发者工具
		},
	})
	if err != nil {
		return fmt.Errorf("启动浏览器失败: %v", err)
	}
	bm.browser = browser

	// 创建页面上下文
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
		Locale: playwright.String("zh-CN"),
	})
	if err != nil {
		return fmt.Errorf("创建浏览器上下文失败: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("创建页面失败: %v", err)
	}
	bm.page = page

	// 注入反检测脚本
	err = bm.injectAntiDetectionScript()
	if err != nil {
		AppLog.Info(fmt.Sprintf("注入反检测脚本失败: %v", err))
	}

	return nil
}

func (bm *BrowserManager) injectAntiDetectionScript() error {
	script := `
		// 移除 webdriver 标识
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined,
		});

		// 伪造 plugins
		Object.defineProperty(navigator, 'plugins', {
			get: () => [1, 2, 3, 4, 5],
		});

		// 伪造 languages
		Object.defineProperty(navigator, 'languages', {
			get: () => ['zh-CN', 'zh', 'en'],
		});

		// 移除自动化相关属性
		delete window.cdc_adoQpoasnfa76pfcZLmcfl_Array;
		delete window.cdc_adoQpoasnfa76pfcZLmcfl_Promise;
		delete window.cdc_adoQpoasnfa76pfcZLmcfl_Symbol;

		// 对抗 disable-devtool
		window.addEventListener('beforeunload', e => e.stopImmediatePropagation(), true);
		window.addEventListener('unload', e => e.stopImmediatePropagation(), true);
		
		// 禁用 disable-devtool 的检测函数
		Object.defineProperty(window, 'outerHeight', { get: () => window.innerHeight });
		Object.defineProperty(window, 'outerWidth', { get: () => window.innerWidth });
		
		// 拦截 disable-devtool 相关方法
		const originalSetInterval = window.setInterval;
		window.setInterval = function(fn, delay) {
			if (delay < 100) return; // 阻止高频检测
			return originalSetInterval.call(this, fn, delay);
		};
		
		// 阻止页面关闭和刷新
		const originalClose = window.close;
		window.close = function() { return false; };
		
		// 完全阻止页面跳转和重定向
		const originalReplace = window.location.replace;
		const originalAssign = window.location.assign;
		const originalReload = window.location.reload;
		
		window.location.replace = function() { console.log('阻止 replace 跳转'); return false; };
		window.location.assign = function() { console.log('阻止 assign 跳转'); return false; };
		window.location.reload = function() { console.log('阻止页面刷新'); return false; };
		
		// 拦截 href 设置
		Object.defineProperty(window.location, 'href', {
			set: function(url) {
				console.log('阻止 href 跳转到:', url);
				return false;
			},
			get: function() {
				return window.location.toString();
			}
		});
		
		// 拦截 history API
		const originalPushState = history.pushState;
		const originalReplaceState = history.replaceState;
		
		history.pushState = function() { console.log('阻止 pushState'); return false; };
		history.replaceState = function() { console.log('阻止 replaceState'); return false; };

		// 完全禁用 disable-devtool 的所有检测机制
		setTimeout(() => {
			// 查找并移除 disable-devtool 相关元素
			const scripts = document.querySelectorAll('script');
			scripts.forEach(script => {
				if (script.src && (script.src.includes('disable-devtool') || script.src.includes('theajack'))) {
					console.log('移除脚本:', script.src);
					script.remove();
				}
				if (script.textContent && script.textContent.includes('disable-devtool')) {
					console.log('移除内联脚本');
					script.remove();
				}
			});
			
			// 清除可能的检测定时器
			for (let i = 1; i < 99999; i++) {
				window.clearInterval(i);
				window.clearTimeout(i);
			}
			
			// 移除可能的事件监听器
			window.removeEventListener('beforeunload', null);
			window.removeEventListener('unload', null);
			document.removeEventListener('visibilitychange', null);
			document.removeEventListener('keydown', null);
			document.removeEventListener('contextmenu', null);
		}, 10);
		
		// 持续监控和清理
		setInterval(() => {
			// 持续清理新添加的脚本
			const newScripts = document.querySelectorAll('script');
			newScripts.forEach(script => {
				if (script.src && (script.src.includes('disable-devtool') || script.src.includes('theajack'))) {
					script.remove();
				}
			});
		}, 1000);

		// 强制启用开发者工具
		document.addEventListener('keydown', e => {
			if (e.key === 'F12') {
				e.stopPropagation();
				e.preventDefault();
				// 强制打开开发者工具
				try { console.clear(); } catch(e) {}
			}
		}, true);
		
		// 覆盖 disable-devtool 的全局变量
		window.DisableDevtool = undefined;
		window.dd = undefined;
		window.disableDevtool = undefined;
		
		// 设置 tkName 跳过检测（根据文档）
		window.ddtk = 'skip';
		window.DDTK = 'skip';
		localStorage.setItem('ddtk', 'skip');
		sessionStorage.setItem('ddtk', 'skip');
		
		// 拦截 localStorage 和 sessionStorage 的 getItem
		const originalGetItem = Storage.prototype.getItem;
		Storage.prototype.getItem = function(key) {
			if (key === 'ddtk' || key === 'DDTK') {
				return 'skip';
			}
			return originalGetItem.call(this, key);
		};
		
		// 覆盖 DisableDevtool 构造函数
		window.DisableDevtool = function() { return { isRunning: false }; };
		window.DisableDevtool.isRunning = false;
		
		// 阻止所有可能的跳转方式
		window.addEventListener('beforeunload', e => {
			e.preventDefault();
			e.stopImmediatePropagation();
			return false;
		}, true);
		
		window.addEventListener('unload', e => {
			e.preventDefault();
			e.stopImmediatePropagation();
			return false;
		}, true);
		
		// 拦截所有可能的跳转事件
		document.addEventListener('click', e => {
			if (e.target.tagName === 'A' && e.target.href && e.target.href.includes('404.html')) {
				e.preventDefault();
				e.stopPropagation();
				console.log('阻止跳转到 404 页面');
				return false;
			}
		}, true);
	`

	err := bm.page.AddInitScript(playwright.Script{Content: playwright.String(script)})
	return err
}

func (bm *BrowserManager) NavigateToURL(url string) error {
	if bm.page == nil {
		return fmt.Errorf("浏览器页面未初始化")
	}

	// 随机延迟模拟人类行为
	time.Sleep(time.Duration(1+time.Now().UnixNano()%3) * time.Second)

	_, err := bm.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(30000),
	})
	return err
}

func (bm *BrowserManager) ExtractDownloadLinks(selector string) ([]string, error) {
	if bm.page == nil {
		return nil, fmt.Errorf("浏览器页面未初始化")
	}

	// 等待页面加载完成
	err := bm.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	if err != nil {
		return nil, fmt.Errorf("等待页面加载失败: %v", err)
	}

	// 如果没有指定选择器，使用通用下载链接选择器
	if selector == "" {
		selector = `a[href*="download"], a[href$=".pdf"], a[href$=".doc"], a[href$=".docx"], a[href$=".zip"], a[href$=".rar"]`
	}

	elements, err := bm.page.QuerySelectorAll(selector)
	if err != nil {
		return nil, fmt.Errorf("查找下载链接失败: %v", err)
	}

	var links []string
	for _, element := range elements {
		href, err := element.GetAttribute("href")
		if err != nil {
			continue
		}
		if href != "" {
			// 处理相对链接
			if strings.HasPrefix(href, "/") {
				currentURL := bm.page.URL()
				href = strings.TrimSuffix(currentURL, "/") + href
			}
			links = append(links, href)
		}
	}

	return links, nil
}

func (bm *BrowserManager) NextPage(nextPageSelector string) (bool, error) {
	if bm.page == nil {
		return false, fmt.Errorf("浏览器页面未初始化")
	}

	// 默认下一页选择器
	if nextPageSelector == "" {
		nextPageSelector = `a:has-text("下一页"), a:has-text("Next"), a:has-text("›"), .next, .pagination-next`
	}

	// 查找下一页按钮
	nextButton, err := bm.page.QuerySelector(nextPageSelector)
	if err != nil || nextButton == nil {
		return false, nil // 没有下一页
	}

	// 检查按钮是否可点击
	isDisabled, _ := nextButton.GetAttribute("disabled")
	if isDisabled != "" {
		return false, nil
	}

	// 模拟人类点击行为
	time.Sleep(time.Duration(1+time.Now().UnixNano()%2) * time.Second)

	err = nextButton.Click()
	if err != nil {
		return false, fmt.Errorf("点击下一页失败: %v", err)
	}

	// 等待页面加载
	err = bm.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	if err != nil {
		return false, fmt.Errorf("等待页面加载失败: %v", err)
	}

	return true, nil
}

func (bm *BrowserManager) Close() error {
	if bm.browser != nil {
		err := bm.browser.Close()
		if err != nil {
			return err
		}
	}
	if bm.pw != nil {
		err := bm.pw.Stop()
		if err != nil {
			return err
		}
	}
	return nil
}
