package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"regexp"
	"strings"
)

var (
	powerAdminErrorDesc = prometheus.NewDesc("poweradmin_error", "Error collecting metrics", nil, nil)
	invalidMetricChars  = regexp.MustCompile("[^a-zA-Z0-9_:]")
)

// Collector generic collector type
type Collector struct {
	PowerAdminClient PAExternalAPI
	Config           Config
}

// NewCollector returns the collector
func NewCollector(client PAExternalAPI, config Config) *Collector {
	return &Collector{PowerAdminClient: client, Config: config}
}

// Describe to satisfy the collector interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Collect metrics from PowerAdmin external API
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	metrics, err := c.PowerAdminClient.GetResources(c.Config.Groups)
	if err != nil {
		log.Infof("Failed to get metrics for groups: %v", err)
		ch <- prometheus.NewInvalidMetric(powerAdminErrorDesc, err)
		return
	}
	log.Infof("Received %d metrics", len(metrics.Values))
	for _, metric := range metrics.Values {
		metricName := getFormattedMetricName(metric.MonitorTitle)
		labels := make(map[string]string, 2)
		labels["group_path"] = metric.GroupPath
		labels["server_name"] = metric.ServerName
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(metricName, metricName, nil, labels),
			prometheus.UntypedValue,
			getFloatValue(metric.MonitorValue, c.Config),
		)
	}
}

func getFormattedMetricName(name string) string {
	// Ensure metric names conform to Prometheus metric name conventions
	metricName := strings.ReplaceAll(name, " ", "_")
	metricName = strings.ToLower(metricName + "_status")
	metricName = strings.ReplaceAll(metricName, "/", "_per_")
	metricName = invalidMetricChars.ReplaceAllString(metricName, "_")
	// metric name cannot start with a digit
	if metricName[0] >= '0' && metricName[0] <= '9' {
		metricName = "_" + metricName
	}
	return metricName
}

func getFloatValue(value string, config Config) float64 {
	if st, stExist := config.StatusMapping.Statuses[strings.ToLower(value)]; stExist {
		return st
	}
	return config.StatusMapping.Default
}
