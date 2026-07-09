// Command lottery 是大乐透助手的入口程序。
package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"gioui.org/app"

	"github.com/user/lottery/client"
	"github.com/user/lottery/service"
	"github.com/user/lottery/ui"
)

func main() {
	// 初始化结构化日志
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("大乐透助手启动中...")

	// 初始化体彩 API 客户端
	sportClient := client.NewSportteryClient()

	// 组装业务服务（DI）
	lotterySvc := service.NewLotteryService(sportClient)
	statsSvc := service.NewStatsService(sportClient)
	recommSvc := service.NewRecommendService(sportClient)

	// 初始化 UI
	a := ui.NewApp(lotterySvc, statsSvc, recommSvc)

	// 在 goroutine 中运行事件循环
	go func() {
		if err := a.Run(); err != nil {
			slog.Error("应用异常退出", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	// 监听系统信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("大乐透助手已启动")

	// app.Main() 在桌面平台启动主循环，阻塞直到窗口关闭
	app.Main()
}
