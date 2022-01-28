package main

import (
	"context"
	"fmt"

	"github.com/gogf/gf/container/gset"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/v2/encoding/ghtml"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

var testContent = `
<pre class="language-go"><code class="language-go">package main

import (
	&quot;github.com/gogf/gf/v2/frame/g&quot;
	&quot;github.com/gogf/gf/v2/net/ghttp&quot;
)

// 优先调用的HOOK
func beforeServeHook1(r *ghttp.Request) {
	r.SetParam(&quot;name&quot;, &quot;GoFrame&quot;)
	r.Response.Writeln(&quot;set name&quot;)
}

// 随后调用的HOOK
func beforeServeHook2(r *ghttp.Request) {
	r.SetParam(&quot;site&quot;, &quot;https://goframe.org&quot;)
	r.Response.Writeln(&quot;set site&quot;)
}

// 允许对同一个路由同一个事件注册多个回调函数，按照注册顺序进行优先级调用。
// 为便于在路由表中对比查看优先级，这里讲HOOK回调函数单独定义为了两个函数。
func main() {
	s := g.Server()
	s.BindHandler(&quot;/&quot;, func(r *ghttp.Request) {
		r.Response.Writeln(r.Get(&quot;name&quot;))
		r.Response.Writeln(r.Get(&quot;site&quot;))
	})
	s.BindHookHandler(&quot;/&quot;, ghttp.HookBeforeServe, beforeServeHook1)
	s.BindHookHandler(&quot;/&quot;, ghttp.HookBeforeServe, beforeServeHook2)
	s.SetPort(8199)
	s.Run()
}
</code></pre>
<p>执行后，终端输出的路由表信息如下：</p>
<pre class="language-undefined"><code class="language-undefined">  SERVER  | ADDRESS | DOMAIN  | METHOD | P | ROUTE |        HANDLER        |    MIDDLEWARE
|---------|---------|---------|--------|---|-------|-----------------------|-------------------|
  default |  :8199  | default | ALL    | 1 | /     | main.main.func1       |
|---------|---------|---------|--------|---|-------|-----------------------|-------------------|
  default |  :8199  | default | ALL    | 2 | /     | main.beforeServeHook1 | HOOK_BEFORE_SERVE
|---------|---------|---------|--------|---|-------|-----------------------|-------------------|
  default |  :8199  | default | ALL    | 1 | /     | main.beforeServeHook2 | HOOK_BEFORE_SERVE
|---------|---------|---------|--------|---|-------|-----------------------|-------------------|
</code>
</pre>
`

func replaceGoDocToPkgGoDev(content string) string {
	return gstr.Replace(content, `godoc.org`, `pkg.go.dev`)
}

func replaceContentCodeToMicro(content string) string {
	newContent, err := gregex.ReplaceStringFuncMatch(
		`<pre\s+class="language\-(\w+)">\s*<code.+?>([\s\S]+?)</code>\s*</pre>`,
		content,
		func(match []string) string {
			var (
				language = match[1]
				codeContent = ghtml.EntitiesDecode(match[2])
			)
			if language == "undefined" {
				language = `xml`
			}
			replacedContent := fmt.Sprintf(gstr.Trim(`
<ac:structured-macro ac:name="code" ac:schema-version="1">
<ac:parameter ac:name="language">%s</ac:parameter>
<ac:plain-text-body><![CDATA[%s]]></ac:plain-text-body>
</ac:structured-macro>
`,), language, gstr.Trim(codeContent))
			return replacedContent
		},
	)
	if err != nil {
		panic(err)
	}
	return newContent
}

func replaceContentToV2(content string) string {
	newContent, err := gregex.ReplaceStringFuncMatch(
		`(github\.com\/gogf\/gf\/)([\/\w]+)`,
		content,
		func(match []string) string {
			ignorePrefixes := g.SliceStr{
				"v2", "issues", "pulls", "actions", "projects", "graphs",
			}
			for _, prefix := range ignorePrefixes {
				if gstr.HasPrefix(match[2], prefix) {
					return match[0]
				}
			}
			return match[1]+"v2/"+match[2]
		},
	)
	if err != nil {
		panic(err)
	}
	return newContent
}

func getContentIds(ctx context.Context) []int{
	sql := `
SELECT DISTINCT CONTENTID FROM CONTENT 
WHERE CONTENTTYPE='PAGE' and CONTENT_STATUS='current' and SPACEID=1310721
`
	list, err := g.DB().GetAll(ctx, sql)
	if err != nil {
		panic(err)
	}
	contentIdSet := gset.NewIntSet()
	for _, item := range list {
		contentIdSet.Add(item["CONTENTID"].Int())
	}
	return contentIdSet.Slice()
}

func replaceContentByIdArray(ctx context.Context, ids []int) {
	sql := fmt.Sprintf(`SELECT * FROM BODYCONTENT WHERE CONTENTID IN(%s)`, gstr.JoinAny(ids, ","))
	list, err := g.DB().GetAll(ctx, sql)
	if err != nil {
		panic(err)
	}
	var content string
	for _, item := range list {
		fmt.Println("replace content id:", item["CONTENTID"].Int())
		content = replaceContentToV2(item["BODY"].String())
		content = replaceContentCodeToMicro(content)
		content = replaceGoDocToPkgGoDev(content)
		_, err = g.Model(`BODYCONTENT`).Data(`BODY`, content).Where(`BODYCONTENTID`, item["BODYCONTENTID"].Int()).Update()
		if err != nil {
			g.Log().Fatal(ctx, err)
		}
	}
}

func main() {
	ctx := gctx.New()
	ids := getContentIds(ctx)
	replaceContentByIdArray(ctx, ids)
}
