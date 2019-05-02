package main

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockPAExternalAPI struct {
	mock.Mock
}

func (mock *MockPAExternalAPI) GetResources(groupName string) (*MonitoredValues, error) {
	args := mock.Called(groupName)
	return args.Get(0).(*MonitoredValues), args.Error(1)
}
func (mock *MockPAExternalAPI) GetMonitorInfos(cid string) (*MonitorInfos, error) {
	args := mock.Called(cid)
	return args.Get(0).(*MonitorInfos), args.Error(1)
}
func (mock *MockPAExternalAPI) GetGroupList() (*GroupList, error) {
	args := mock.Called()
	return args.Get(0).(*GroupList), args.Error(1)
}
func (mock *MockPAExternalAPI) GetServerList(gid string) (*ServerList, error) {
	args := mock.Called(gid)
	return args.Get(0).(*ServerList), args.Error(1)
}
func TestCollector_Collect(t *testing.T) {
	api := MockPAExternalAPI{}
	collector := NewCollector(&api)
	values := make([]MonitoredValue, 1)
	value := MonitoredValue{
		MonitorTitle:   "Toto",
		MonitorValue:   "OK",
		MonitorLastRun: time.Now(),
		MonitorStatus:  "OK",
		ServerID:       "158",
		GroupID:        "154",
	}
	values = append(values, value)
	m := MonitoredValues{
		values,
	}
	api.On("GetResources").Return(m)
	ch := make(chan prometheus.Metric)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	for m := range ch {
		got := readMetric(m)
		assert.Equal(t, dto.MetricType_UNTYPED, got.metricType)
		assert.Equal(t, 1, got.value)
	}

}

type labelMap map[string]string

type MetricResult struct {
	labels     labelMap
	value      float64
	metricType dto.MetricType
}

func readMetric(m prometheus.Metric) MetricResult {
	pb := &dto.Metric{}
	m.Write(pb)
	labels := make(labelMap, len(pb.Label))
	for _, v := range pb.Label {
		labels[v.GetName()] = v.GetValue()
	}
	if pb.Gauge != nil {
		return MetricResult{labels: labels, value: pb.GetGauge().GetValue(), metricType: dto.MetricType_GAUGE}
	}
	if pb.Counter != nil {
		return MetricResult{labels: labels, value: pb.GetCounter().GetValue(), metricType: dto.MetricType_COUNTER}
	}
	if pb.Untyped != nil {
		return MetricResult{labels: labels, value: pb.GetUntyped().GetValue(), metricType: dto.MetricType_UNTYPED}
	}
	panic("Unsupported metric type")
}
