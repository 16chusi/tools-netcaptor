package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

type ChromeDPManager struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewChromeDPManager() *ChromeDPManager {
	return &ChromeDPManager{}
}

func (c *ChromeDPManager) Init() error {
	// 创建 Chrome 选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("exclude-switches", "enable-automation"),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(format string, args ...interface{}) {
		AppLog.Info(fmt.Sprintf(format, args...))
	}))

	c.ctx = ctx
	c.cancel = cancel

	// 启动浏览器
	err := chromedp.Run(ctx)
	if err != nil {
		return fmt.Errorf("启动 ChromeDP 失败: %v", err)
	}

	return nil
}

func (c *ChromeDPManager) NavigateToURL(url string) error {
	if c.ctx == nil {
		return fmt.Errorf("浏览器未初始化")
	}

	// 注入反检测脚本
	antiDetectScript := `
		// 移除 webdriver 标识
		Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
		
		// 伪造 plugins
		Object.defineProperty(navigator, 'plugins', { get: () => [1, 2, 3, 4, 5] });
		
		// 伪造 languages
		Object.defineProperty(navigator, 'languages', { get: () => ['zh-CN', 'zh', 'en'] });
		
		// 移除自动化相关属性
		delete window.cdc_adoQpoasnfa76pfcZLmcfl_Array;
		delete window.cdc_adoQpoasnfa76pfcZLmcfl_Promise;
		delete window.cdc_adoQpoasnfa76pfcZLmcfl_Symbol;
		
		// 对抗 disable-devtool
		Object.defineProperty(window, 'outerHeight', { get: () => window.innerHeight });
		Object.defineProperty(window, 'outerWidth', { get: () => window.innerWidth });
		
		// 阻止页面跳转
		window.location.replace = function() { console.log('阻止 replace'); return false; };
		window.location.assign = function() { console.log('阻止 assign'); return false; };
		window.close = function() { console.log('阻止关闭'); return false; };
		
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
		
		// 清除 disable-devtool 脚本
		setTimeout(() => {
			const scripts = document.querySelectorAll('script');
			scripts.forEach(script => {
				if (script.src && (script.src.includes('disable-devtool') || script.src.includes('theajack'))) {
					script.remove();
				}
			});
		}, 100);
	`

	err := chromedp.Run(c.ctx,
		chromedp.Navigate(url),
		chromedp.Evaluate(antiDetectScript, nil),
		chromedp.WaitReady("body"),
	)

	if err != nil {
		return fmt.Errorf("访问页面失败: %v", err)
	}

	return nil
}

func (c *ChromeDPManager) ExtractDownloadLinks(selector string) ([]string, error) {
	if c.ctx == nil {
		return nil, fmt.Errorf("浏览器未初始化")
	}

	// 默认选择器
	if selector == "" {
		selector = `a[href*="download"], a[href$=".pdf"], a[href$=".doc"], a[href$=".docx"], a[href$=".zip"], a[href$=".rar"]`
	}

	var links []string
	err := chromedp.Run(c.ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			Array.from(document.querySelectorAll('%s')).map(a => {
				let href = a.href;
				if (href.startsWith('/')) {
					href = window.location.origin + href;
				}
				return href;
			}).filter(href => href && href !== '');
		`, selector), &links),
	)

	if err != nil {
		return nil, fmt.Errorf("提取链接失败: %v", err)
	}

	return links, nil
}

func (c *ChromeDPManager) NextPage(nextPageSelector string) (bool, error) {
	if c.ctx == nil {
		return false, fmt.Errorf("浏览器未初始化")
	}

	// 默认下一页选择器
	if nextPageSelector == "" {
		nextPageSelector = `a:contains("下一页"), a:contains("Next"), a:contains("›"), .next, .pagination-next`
	}

	// 检查下一页按钮是否存在
	var exists bool
	err := chromedp.Run(c.ctx,
		chromedp.Evaluate(fmt.Sprintf(`
			const nextBtn = document.querySelector('%s');
			nextBtn && !nextBtn.disabled && nextBtn.style.display !== 'none';
		`, nextPageSelector), &exists),
	)

	if err != nil || !exists {
		return false, nil
	}

	// 点击下一页
	err = chromedp.Run(c.ctx,
		chromedp.Click(nextPageSelector),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitReady("body"),
	)

	if err != nil {
		return false, fmt.Errorf("点击下一页失败: %v", err)
	}

	return true, nil
}

func (c *ChromeDPManager) GetCurrentURL() (string, error) {
	if c.ctx == nil {
		return "", fmt.Errorf("浏览器未初始化")
	}

	var url string
	err := chromedp.Run(c.ctx,
		chromedp.Evaluate(`window.location.href`, &url),
	)

	return url, err
}

func (c *ChromeDPManager) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}
