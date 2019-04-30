package main

import (
	"github.com/FXInnovation/poweradmin_exporter/poweradmin"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"regexp"
	"strings"
)

var (
	powerAdminErrorDesc = prometheus.NewDesc("poweradmin_error", "Error collecting metrics", nil, nil)
	invalidMetricChars  = regexp.MustCompile("[^a-zA-Z0-9_:]")
)

// Collector generic collector type
type Collector struct{}

// NewCollector returns the collector
func NewCollector() *Collector {
	return &Collector{}
}

// Describe to satisfy the collector interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Collect metrics from ...
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	pc, _ := poweradmin.NewPAExternalAPIClient(config.APIKey, config.ServerURL)

	for _, group := range config.Groups {
		groupName := group.GroupName
		metrics, err := pc.GetResources(groupName)
		if err != nil {
			log.Printf("Failed to get metrics for group %s: %v", groupName, err)
			ch <- prometheus.NewInvalidMetric(powerAdminErrorDesc, err)
			return
		}
		for _, metric := range metrics.Values {
			metricName := getMetricName(metric.MonitorTitle)
			log.Printf("Metric sent %s:%s", metricName, metric)
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(metricName, metricName, nil, nil),
				prometheus.UntypedValue,
				getFloatValue(metric.MonitorValue),
			)
		}
	}
}

func getMetricName(name string) string {
	// Ensure metric names conform to Prometheus metric name conventions
	metricName := strings.Replace(name, " ", "_", -1)
	metricName = strings.ToLower(metricName + "_status")
	metricName = strings.Replace(metricName, "/", "_per_", -1)
	metricName = invalidMetricChars.ReplaceAllString(metricName, "_")
	return metricName
}

func getFloatValue(value string) float64 {
	if strings.ToLower(value) == "ok" {
		return 1
	}
	return 0
}
