package main

import (
	"fmt"
	"github.com/gogf/gf/container/garray"
	"github.com/gogf/gf/frame/g"
)

func main() {
	sql := `
SELECT DISTINCT CONTENTID FROM CONTENT 
WHERE CONTENTTYPE='PAGE' and CONTENT_STATUS='current'  and SPACEID > 0
`
	list, err := g.DB().GetAll(sql)
	if err != nil {
		panic(err)
	}
	urlArray := garray.NewStrArray()
	for _, item := range list {
		urlArray.Append(fmt.Sprintf(`https://itician.org/pages/viewpage.action?pageId=%d`, item["CONTENTID"].Int()))
	}
	response := g.Client().ContentType(`text/plain`).PostContent(
		`http://data.zz.baidu.com/urls?site=https://itician.org&token=`+g.Cfg().GetString("baidu.ziyuan.token"),
		urlArray.Join("\n"),
	)
	fmt.Println(response)
}
