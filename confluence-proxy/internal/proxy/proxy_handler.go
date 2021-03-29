package proxy

import (
	"fmt"
	"github.com/gogf/gf/container/gset"
	"github.com/gogf/gf/encoding/gcompress"
	"github.com/gogf/gf/encoding/ghtml"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"net/http"
)

var (
	staticFileExtSetForCdn = gset.NewStrSetFrom([]string{
		// 样式文件
		"map", "less", "sass", "js", "json", "css",
		// 网页文件
		"xml", "htm", "html", "xhtml", "shtml", "tpl",
		// 图片文件
		"png", "gif", "svg", "jpg", "jpeg", "bmp", "ico",
		// 字体文件
		"woff", "woff2", "ttf", "eot",
		// 压缩文件
		"zip", "rar", "7z", "gz", "bz2",
		// 文档文件
		"doc", "docx", "pdf", "xls", "xlsx", "ppt", "txt", "log", "psd", "md",
	})

	cdnStaticFileTypesStr = gstr.Join(staticFileExtSetForCdn.Slice(), "|")
)

// 处理返回
func requestBeforeProxyHandler(r *http.Request) {

}

// 处理返回
func responseHandler(writer *ResponseWriter, r *http.Request) {
	var (
		responseBody = writer.BufferString()
	)
	// 解析并清除压缩
	if gstr.Equal(writer.Header().Get("Content-Encoding"), "gzip") {
		writer.Header().Del("Content-Encoding")
		content, _ := gcompress.UnGzip(writer.Buffer())
		responseBody = string(content)
	}
	// 增加keywords/description
	var (
		keywordsMeta    string
		descriptionMeta string
	)
	// SEO标题 - keywords, description
	keywordsMeta = fmt.Sprintf(`<meta name="keywords" content="%s" />`, g.Cfg().GetString("site.keywords"))
	if gstr.Contains(responseBody, `class="wiki-content"`) {
		descriptionMeta, _ = gregex.ReplaceString(`[\s\S]+(<div.+?class="wiki\-content")`, `$1`, responseBody)
		descriptionMeta = ghtml.StripTags(descriptionMeta)
		descriptionMeta, _ = gregex.ReplaceString(`<.+?>`, ``, descriptionMeta)
		descriptionMeta, _ = gregex.ReplaceString(`\s+`, ` `, descriptionMeta)
		descriptionMeta = gstr.SubStrRune(descriptionMeta, 0, 360)
		descriptionMeta = gstr.Trim(descriptionMeta)
		descriptionMeta = fmt.Sprintf(`<meta name="description" content="%s" />`, descriptionMeta)
	} else {
		descriptionMeta = fmt.Sprintf(`<meta name="description" content="%s" />`, g.Cfg().GetString("site.description"))
	}
	// SEO标题
	responseBody = gstr.Replace(responseBody, `主页面 - `, ``, 1)
	responseBody = gstr.Replace(responseBody, `Confluence Mobile - `, ``, 1)
	responseBody = gstr.Replace(responseBody, ` - Dashboard - `, ` - `, 1)
	responseBody = gstr.Replace(responseBody, ` - GoFrame (ZH) - `, ` - `, 1)
	responseBody = gstr.Replace(responseBody, `</title>`, `</title>`+keywordsMeta+descriptionMeta, 1)

	// CDN处理
	if gstr.Contains(r.URL.Path, "/viewpage.action") || gstr.Contains(r.URL.Path, "/display/") {
		// 文章受限访问
		if !gstr.Contains(responseBody, "aui-iconfont-locked") && !gstr.Contains(responseBody, "aui-iconfont-unlocked restricted") {
			responseBody = handleContentReplacementForCDN(responseBody)
		}
	}

	// 去掉robots，允许搜索引擎收录
	responseBody = gstr.Replace(responseBody, `<meta name="robots"`, `<meta name="no-robots"`, 1)
	writer.OverwriteString(responseBody)
}

// CDN连接处理
func handleContentReplacementForCDN(responseBody string) string {
	// HTML CSS/JS
	responseBody, _ = gregex.ReplaceString(
		fmt.Sprintf(`(src|href)="/([^"']+\.(%s)[^"']*)"`, cdnStaticFileTypesStr),
		`$1="https://gfcdn.johng.cn/$2"`,
		responseBody,
	)
	// CSS URL
	//responseBody, _ = gregex.ReplaceString(
	//	fmt.Sprintf(`url\(/(.+\.(%s).*)\)`, cdnStaticFileTypesStr),
	//	`url(/https://gfcdn.johng.cn/$1"`,
	//	responseBody,
	//)
	return responseBody
}
