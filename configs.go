package main

// SiteConfig 网站配置结构
type SiteConfig struct {
	Name             string `json:"name"`
	Domain           string `json:"domain"`
	LinkSelector     string `json:"linkSelector"`
	NextPageSelector string `json:"nextPageSelector"`
	Description      string `json:"description"`
}

// GetPresetConfigs 获取预设的网站配置
func (a *App) GetPresetConfigs() []SiteConfig {
	return []SiteConfig{
		{
			Name:             "通用文档下载",
			Domain:           "*",
			LinkSelector:     `a[href$=".pdf"], a[href$=".doc"], a[href$=".docx"], a[href*="download"]`,
			NextPageSelector: `a:has-text("下一页"), a:has-text("Next"), .next, .pagination-next`,
			Description:      "适用于大多数文档下载网站",
		},
		{
			Name:             "压缩包下载",
			Domain:           "*",
			LinkSelector:     `a[href$=".zip"], a[href$=".rar"], a[href$=".7z"], a[href$=".tar.gz"]`,
			NextPageSelector: `a:has-text("下一页"), a:has-text("Next"), .next`,
			Description:      "专门用于压缩包文件下载",
		},
		{
			Name:             "学术论文",
			Domain:           "*.edu",
			LinkSelector:     `a[href$=".pdf"], .pdf-link, .download-pdf`,
			NextPageSelector: `.page-next, a[title*="next"], .pagination .next`,
			Description:      "适用于学术网站的论文下载",
		},
		{
			Name:             "软件下载",
			Domain:           "*",
			LinkSelector:     `a[href$=".exe"], a[href$=".msi"], a[href$=".dmg"], a[href$=".deb"], .download-btn`,
			NextPageSelector: `.next-page, a:has-text("下一页")`,
			Description:      "软件和应用程序下载",
		},
		{
			Name:             "媒体文件",
			Domain:           "*",
			LinkSelector:     `a[href$=".mp4"], a[href$=".mp3"], a[href$=".avi"], a[href$=".mkv"]`,
			NextPageSelector: `.pagination .next, a:has-text("›")`,
			Description:      "音频和视频文件下载",
		},
		{
			Name:             "GitHub Releases",
			Domain:           "github.com",
			LinkSelector:     `a[href*="/releases/download/"]`,
			NextPageSelector: `a[rel="next"]`,
			Description:      "GitHub 项目发布文件",
		},
	}
}

// ApplySiteConfig 应用网站配置
func (a *App) ApplySiteConfig(config SiteConfig) error {
	// 这个方法可以在前端调用来应用预设配置
	// 实际的配置应用在前端完成
	return nil
}
