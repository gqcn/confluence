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
func responseHandler(writer *ResponseWriter) {
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
	descriptionMeta, _ = gregex.ReplaceString(`[\s\S]+(<div\s+id="main\-content")`, `$1`, responseBody)
	descriptionMeta = ghtml.StripTags(descriptionMeta)
	descriptionMeta = gstr.TrimLeft(descriptionMeta)
	descriptionMeta = gstr.SubStrRune(descriptionMeta, 0, 360)
	descriptionMeta = fmt.Sprintf(`<meta name="description" content="%s" />`, descriptionMeta)
	responseBody = gstr.ReplaceByMap(responseBody, g.MapStrStr{
		`</title>`: `</title>` + keywordsMeta + descriptionMeta,
		`<meta name="robots" content="noindex,nofollow">`: "",
		`<meta name="robots" content="noarchive">`:        "",
	})
	writer.OverwriteString(responseBody)
}
