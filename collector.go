package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

var (
	powerAdminErrorDesc = prometheus.NewDesc("poweradmin_error", "Error collecting metrics", nil, nil)
	invalidMetricChars  = regexp.MustCompile("[^a-zA-Z0-9_: ]")
	statusConfig        StatusConfig
)

// Collector generic collector type
type Collector struct {
	PowerAdminClient PAExternalAPI
}

func init() {
	statusConfig = loadStatuses("status_mapping.yml")
}

// NewCollector returns the collector
func NewCollector(client PAExternalAPI) *Collector {
	return &Collector{PowerAdminClient: client}
}

// Describe to satisfy the collector interface.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Collect metrics from PowerAdmin external API
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	metrics, err := c.PowerAdminClient.GetResources(config.Groups)
	if err != nil {
		log.Printf("Failed to get metrics for groups: %v", err)
		ch <- prometheus.NewInvalidMetric(powerAdminErrorDesc, err)
		return
	}
	for _, metric := range metrics.Values {
		metricName := getFormattedMetricName(metric.MonitorTitle)
		log.Printf("Metric sent %s:%s", metricName, metric)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(metricName, metricName, nil, nil),
			prometheus.UntypedValue,
			getFloatValue(metric.MonitorValue),
		)
	}

}

func getFormattedMetricName(name string) string {
	// Ensure metric names conform to Prometheus metric name conventions
	metricName := strings.ReplaceAll(name, " ", "_")
	metricName = strings.ToLower(metricName + "_status")
	metricName = strings.ReplaceAll(metricName, "/", "_per_")
	metricName = invalidMetricChars.ReplaceAllString(metricName, "_")
	return metricName
}

func getFloatValue(value string) float64 {
	if st, stExist := statusConfig.Statuses[strings.ToLower(value)]; stExist {
		return st
	}
	return statusConfig.Default
}

// StatusConfig configure the status values to be sent
type StatusConfig struct {
	Statuses map[string]float64 `yaml:"values"`
	Default  float64            `yaml:"default"`
}

func loadStatuses(filename string) StatusConfig {
	thisConfig := StatusConfig{}
	statusData, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Unable to read status config file")
	}
	yamlErr := yaml.Unmarshal([]byte(statusData), &thisConfig)
	if yamlErr != nil {
		log.Fatal("Problem unmarshalling status config file")
	}
	return thisConfig
}
