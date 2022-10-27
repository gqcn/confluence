
# 目录介绍


## conflunece

confluence使用到的一些资源文件。

## confluence-heart

健康检查，当confluence访问不通或者超时时重启confluence。

## confluence-proxy

前置Nginx会将请求转发到proxy，proxy再请求到真实的confluence服务，并且中间会做些自定义的字符串替换。

## confluence-seo

每日定时自动生成百度收录的文章文件，给百度爬虫用。

## confluence-replace

临时脚本文件，用于自定义批量替换数据库中的内容。

