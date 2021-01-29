package proxy

import (
	"fmt"
	"github.com/gogf/gf/encoding/gcompress"
	"github.com/gogf/gf/encoding/ghtml"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"net/http"
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
	if r.URL.Path == "/" {
		responseBody = gstr.Replace(responseBody, `<title>`, `<title>`+g.Cfg().GetString("site.title")+` - `, 1)
	}
	responseBody = gstr.Replace(responseBody, ` - GoFrame (ZH) - `, ` - `, 1)
	responseBody = gstr.Replace(responseBody, ` - 主页面 - `, ` - `, 1)
	responseBody = gstr.Replace(responseBody, ` - Dashboard - `, ` - `, 1)
	responseBody = gstr.Replace(responseBody, `</title>`, `</title>`+keywordsMeta+descriptionMeta, 1)
	responseBody = gstr.Replace(responseBody, `<meta name="robots" content="noindex,nofollow">`, ``, 1)
	responseBody = gstr.Replace(responseBody, `<meta name="robots" content="noarchive">`, ``, 1)
	writer.OverwriteString(responseBody)
}
