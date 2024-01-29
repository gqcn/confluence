package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/container/gtype"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gproc"
	"github.com/gogf/gf/v2/os/gtimer"
)

type Config struct {
	Address  string
	Startup  string
	Shutdown string
}

const (
	configKey          = `confluence`
	heartBeatFailedMax = 3
)

var (
	ctx                  = gctx.GetInitCtx()
	config               = &Config{}
	heartBeatFailedCount = gtype.NewInt()
)

// gf build -a amd64 -s linux
func main() {
	err := g.Cfg().MustGet(ctx, configKey).Scan(config)
	if err != nil {
		g.Log().Fatal(ctx, err)
	}
	gtimer.AddSingleton(ctx, time.Second*10, heartbeat)
	g.Log().Infof(ctx, `ok, let's start`)
	g.Listen()
}

func heartbeat(ctx context.Context) {
	var ok = false
	res, err := g.Client().Timeout(time.Second*10).Get(ctx, config.Address)
	if err == nil && res != nil {
		if res.StatusCode == http.StatusOK {
			ok = true
			heartBeatFailedCount.Set(0)
		}
	}

	if !ok {
		heartBeatFailedCount.Add(1)
	}
	g.Log().Infof(ctx, `heartbeat ok: %v, failed count: %d`, ok, heartBeatFailedCount.Val())
	if heartBeatFailedCount.Val() >= heartBeatFailedMax {
		restartConfluence(ctx)
		heartBeatFailedCount.Set(0)
	}
}

func restartConfluence(ctx context.Context) {
	g.Log().Infof(ctx, `restartConfluence start`)
	defer g.Log().Infof(ctx, `restartConfluence end`)

	// 正常停止confluence
	_ = gproc.ShellRun(ctx, fmt.Sprintf(`bash %s`, config.Shutdown))
	time.Sleep(time.Second * 5)
	_ = gproc.ShellRun(ctx, fmt.Sprintf(`bash %s`, config.Shutdown))
	time.Sleep(time.Second * 5)
	_ = gproc.ShellRun(ctx, fmt.Sprintf(`bash %s`, config.Shutdown))
	time.Sleep(time.Second * 10)

	// 如果上面无法正常停止confluence，这里最后强制kill
	_ = gproc.ShellRun(ctx, `killall -9 /home/john/Softs/confluence/jre//bin/java`)
	_ = gproc.ShellRun(ctx, `killall -9 /home/john/Softs/confluence/jre/bin/java`)
	_ = gproc.ShellRun(ctx, `killall -9 java`)
	time.Sleep(time.Second * 5)

	// 启动confluence，等待一段时间后再重新健康探测
	_ = gproc.ShellRun(ctx, fmt.Sprintf(`bash %s`, config.Startup))
	time.Sleep(time.Minute * 10)
}
