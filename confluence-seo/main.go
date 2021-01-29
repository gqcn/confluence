package main

import (
	"fmt"
	"github.com/gogf/gf/container/garray"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gcron"
	"github.com/gogf/gf/os/gfile"
)

func makeBaiduSiteMap() {
	sql := `
SELECT DISTINCT CONTENTID FROM CONTENT 
WHERE CONTENTTYPE='PAGE' and CONTENT_STATUS='current' and SPACEID > 0
`
	list, err := g.DB().GetAll(sql)
	if err != nil {
		panic(err)
	}
	urlArray := garray.NewStrArray()
	for _, item := range list {
		urlArray.Append(fmt.Sprintf(`https://goframe.org/pages/viewpage.action?pageId=%d`, item["CONTENTID"].Int()))
	}
	gfile.PutContents(gfile.Join(g.Cfg().GetString(`setting.sitemap`), "sitemap.txt"), urlArray.Join("\n"))
}

func makeBaiduApiRequest() {
	sql := `
SELECT DISTINCT CONTENTID FROM CONTENT 
WHERE CONTENTTYPE='PAGE' and CONTENT_STATUS='current' and SPACEID > 0
`
	list, err := g.DB().GetAll(sql)
	if err != nil {
		panic(err)
	}
	urlArray := garray.NewStrArray()
	for _, item := range list {
		urlArray.Append(fmt.Sprintf(`https://goframe.org/pages/viewpage.action?pageId=%d`, item["CONTENTID"].Int()))
	}
	g.Client().ContentType(`text/plain`).PostContent(
		`http://data.zz.baidu.com/urls?site=goframe.org&token=`+g.Cfg().GetString("baidu.ziyuan.token"),
		urlArray.Join("\n"),
	)
}

func main() {
	// 启动的时候执行一次
	makeBaiduSiteMap()
	makeBaiduApiRequest()
	// 随后每天凌晨执行
	gcron.Add(`0 0 0 * * *`, makeBaiduSiteMap)
	gcron.Add(`0 0 0 * * *`, makeBaiduApiRequest)
	select {}
}
