package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Collector generic collector type
type Collector struct {}

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

}
