package main

import (
	"regexp"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	powerAdminErrorDesc   = prometheus.NewDesc("poweradmin_error", "Error collecting metrics", nil, nil)
	invalidMetricChars    = regexp.MustCompile("[^a-zA-Z0-9_:]")
	trailingUnderscoresRe = regexp.MustCompile(`_+$`)
)

// Collector generic collector type
type Collector struct {
	PowerAdminClient PAExternalAPI
	DB               SQLServerInterface
	Config           Config
}

// NewCollector returns the collector
func NewCollector(client PAExternalAPI, db SQLServerInterface, config Config) *Collector {
	return &Collector{PowerAdminClient: client, DB: db, Config: config}
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
	log.Infof("Received %d status metrics", len(metrics.Values))
	for _, metric := range metrics.Values {
		exposeMetric(metric.MonitorTitle+"_status", getFloatValue(metric.MonitorValue, c.Config), metric.GroupPath, metric.ServerName, ch)
	}

	// SECTION collect data from PA database
	err = c.DB.GetConnection()
	if err != nil {
		log.Errorf("Failed to connect to PA database: %s", err.Error())
		ch <- prometheus.NewInvalidMetric(powerAdminErrorDesc, err)
		return
	}

	// take each server from each group, lookup their CompID
	// and retrieve the last entry for all available metrics
	for _, confGrp := range c.Config.Groups {
		serverList := confGrp.Servers

		if len(serverList) == 0 {
			serverList, err = c.DB.GetAllServersFor(confGrp.GroupPath)
		}

		for _, server := range serverList {
			compInfo, err := c.DB.GetConfigComputerInfo(server)
			if err != nil {
				log.Error(err.Error())
				ch <- prometheus.NewInvalidMetric(powerAdminErrorDesc, err)
				return
			}
			dbMetrics, err := c.DB.GetAllServerMetric(compInfo.CompID)
			if err != nil {
				log.Error(err.Error())
				ch <- prometheus.NewInvalidMetric(powerAdminErrorDesc, err)
				return
			}
			// log.Info(dbMetrics)
			log.Infof("Received %d metrics", len(dbMetrics))
			for _, dbm := range dbMetrics {
				exposeMetric(getDBMetricName(dbm), dbm.Value, confGrp.GroupPath, server, ch)
			}
		}
	}

}

func getDBMetricName(serverMetric ServerMetric) string {
	var metricName strings.Builder
	metricName.WriteString(serverMetric.ItemAlias)
	if serverMetric.UnitStr != "" {
		metricName.WriteString("__")
		metricName.WriteString(serverMetric.UnitStr)
	} else {
		log.Infof("Metric %s:%s has no unit str with id %d", serverMetric.ItemName, serverMetric.ItemAlias, serverMetric.Unit)
	}
	return metricName.String()
}

func exposeMetric(originalName string, value float64, group string, server string, ch chan<- prometheus.Metric) {
	metricName := getFormattedMetricName(originalName)
	labels := make(map[string]string, 2)
	labels["group_path"] = group
	labels["server_name"] = server
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(metricName, metricName, nil, labels),
		prometheus.UntypedValue,
		value,
	)
}

func getFormattedMetricName(name string) string {
	// Ensure metric names conform to Prometheus metric name conventions
	metricName := strings.ReplaceAll(name, " ", "_")
	metricName = strings.ToLower(metricName)
	metricName = strings.ReplaceAll(metricName, "/", "_per_")
	metricName = invalidMetricChars.ReplaceAllString(metricName, "_")
	// metric name cannot start with a digit
	if metricName[0] >= '0' && metricName[0] <= '9' {
		metricName = "_" + metricName
	}
	// remove trailing _
	metricName = trailingUnderscoresRe.ReplaceAllString(metricName, "")
	return metricName
}

func getFloatValue(value string, config Config) float64 {
	if st, stExist := config.StatusMapping.Statuses[strings.ToLower(value)]; stExist {
		return st
	}
	return config.StatusMapping.Default
}
