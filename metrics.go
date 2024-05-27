package main

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

const (
	app_name          = "spotify_ingest"
	spotify_api_group = "spotify_api"
)

type SpotifyApiMetrics struct {
	totalRequests prometheus.Counter
}

type MetricHandler struct {
	// ingest
	gatewayUrl        string
	spotifyApiMetrics SpotifyApiMetrics
}

func newSpotifyApiMetrics() SpotifyApiMetrics {
	totalRequests := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "spotify_ingest_spotify_api_requests_total",
		Help: "The amount of requests sent to the spotify API",
	})

	return SpotifyApiMetrics{
		totalRequests: totalRequests,
	}
}

func NewMetricHandler(gatewayUrl string) MetricHandler {
	return MetricHandler{
		gatewayUrl: gatewayUrl,
		// metricsByName: MakeMetrics(),
		spotifyApiMetrics: newSpotifyApiMetrics(),
	}

}

func (m *MetricHandler) Push() error {
	if err := push.New(m.gatewayUrl, "spotify_api").
		Collector(m.spotifyApiMetrics.totalRequests).
		Grouping("spotify", "api").
		Push(); err != nil {
		fmt.Println("Could not push puzzy time to Pushgateway:", err)
		return err
	}

	fmt.Println("Pushy PUSHY YOYOY")
	return errors.New("yee")
}
