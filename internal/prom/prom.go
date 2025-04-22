package prom

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vinted/link-exporter/internal/config"
)

type linkCollector struct {
	packetLoss *prometheus.Desc
	latency    *prometheus.Desc
}

type LinkMetrics struct {
	PacketLoss float64
	Latency    float64
}

var MetricsCtx map[string]LinkMetrics

func NewLinkCollector() *linkCollector {
	for _, iface := range config.Config.Interfaces {
		MetricsCtx = make(map[string]LinkMetrics)
		h := LinkMetrics{
			PacketLoss: 0,
			Latency:    0,
		}
		MetricsCtx[iface] = h
	}
	return &linkCollector{
		packetLoss: prometheus.NewDesc("link_exporter_packet_loss",
			"Number of lost packages",
			[]string{"iface"}, nil,
		),
		latency: prometheus.NewDesc("link_exporter_latency",
			"Latency of a round-trip of a package",
			[]string{"iface"}, nil,
		),
	}
}

func (collector *linkCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.packetLoss
	ch <- collector.latency
}

func (collector *linkCollector) Collect(ch chan<- prometheus.Metric) {
	for _, iface := range config.Config.Interfaces {
		ch <- prometheus.MustNewConstMetric(collector.packetLoss, prometheus.GaugeValue, MetricsCtx[iface].PacketLoss, iface)
		ch <- prometheus.MustNewConstMetric(collector.latency, prometheus.GaugeValue, MetricsCtx[iface].Latency, iface)
	}
}

func PromHTTPStart() {
	srv := http.NewServeMux()
	srv.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(config.Config.HttpListenAddress, srv); err != nil {
		slog.Error(fmt.Sprintf("Unable to start server: %v", err))
	}
}
