package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/vinted/link-exporter/internal/config"
	"github.com/vinted/link-exporter/internal/probe"
	"github.com/vinted/link-exporter/internal/prom"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vinted/link-exporter/internal/reflector"
)

func main() {
	var (
		configPath string
	)
	flag.StringVar(&configPath, "config_file", "/etc/link-exporter/config.json", "Path to config file.")
	flag.Parse()

	err := config.Init(configPath)
	if err != nil {
		slog.Error("failed to read configuration: %s", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevel(),
	}))
	slog.SetDefault(logger)
	lc := prom.NewLinkCollector()
	prometheus.MustRegister(lc)
	go prom.PromHTTPStart()
	go probe.Start()
	reflector.Start()

}

func getLogLevel() slog.Level {
	switch config.Config.LogLevel {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
