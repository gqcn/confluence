package proxy

import (
	"errors"
	"github.com/gogf/gf/container/gset"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/text/gstr"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var (
	server   *http.Server
	address  = g.Cfg().GetString("proxy.address")
	upstream = g.Cfg().GetString("proxy.upstream")
	// 常见静态文件访问不做链路跟踪处理，提高请求转发性能
	staticFileExtSet = gset.NewStrSetFrom([]string{
		// 样式文件
		"map", "less", "sass",
		"js", "json", "css",
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
	// 防盗链允许访问域名
	staticDefenderNoneBlocks = gstr.SplitAndTrim(g.Cfg().GetString("proxy.staticDefenderNoneBlocks"), ",")
)

func init() {
	if address == "" {
		g.Log().Fatal("http proxy address cannot be empty")
	}
	if upstream == "" {
		g.Log().Fatal("http proxy upstream cannot be empty")
	}
}

func Run() {
	server = &http.Server{
		Addr:         address,
		Handler:      http.HandlerFunc(httpHandler),
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		IdleTimeout:  time.Minute,
	}
	// 启动HTTP Server服务
	g.Log().Printf("%d: http proxy start running on %s", gproc.Pid(), address)
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		g.Log().Error(err)
	}
}

// 默认的HTTP反向代理处理方法
func httpHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err error
	)
	// 判断静态文件请求
	isStaticRequest := false
	if ext := gfile.ExtName(r.URL.Path); ext != "" {
		if staticFileExtSet.Contains(ext) {
			isStaticRequest = true
		}
	}
	// 检测反向代理配置，如果不存在则返回404
	if upstream == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	// 防盗链
	if isStaticRequest && defendStealing(w, r) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(http.StatusText(http.StatusForbidden)))
		return
	}

	// 创建自定义的Writer，支持缓存控制
	writer := NewResponseWriter(w)

	// 非静态文件请求才执行内容替换
	if !isStaticRequest {
		// 反向代理日志记录
		defer func() {
			if err != nil {
				g.Log().Error(err)
			}
			responseHandler(writer, r)
			// 将缓存的返回内容输出到客户端
			writer.OutputBuffer()
		}()
		requestBeforeProxyHandler(r)
	}

	// 反向代理请求处理，后端HTTP目标服务统一使用HTTP
	var u *url.URL
	u, err = url.Parse(upstream)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		writer.WriteHeader(http.StatusBadGateway)
		err = e
	}

	if isStaticRequest {
		// 静态文件服务使用底层Writer支持Stream流式下载
		proxy.ServeHTTP(writer.RawWriter(), r)
	} else {
		// 非静态文件请求使用缓存Writer
		proxy.ServeHTTP(writer, r)
	}
}

// 防盗链。如果防盗成功，那么返回true，否则false。
func defendStealing(w http.ResponseWriter, r *http.Request) bool {
	// 关闭防盗链
	return false

	if ext := gfile.ExtName(r.URL.Path); ext != "" {
		switch ext {
		case
			// 样式文件
			"js", "json", "css", "map", "less", "sass",
			// 图片文件
			"png", "gif", "svg", "jpg", "jpeg", "bmp",
			// 字体文件
			"woff", "woff2", "ttf", "eot":
			var (
				referer = r.Referer()
			)
			for _, v := range staticDefenderNoneBlocks {
				if gstr.Contains(referer, v) {
					return false
				}
			}
			return true
		}
	}
	return false
}
