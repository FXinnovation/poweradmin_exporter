package main

import (
	"errors"
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

func (mock *MockPAExternalAPI) GetResources(groupFilters []GroupFilter) (*MonitoredValues, error) {
	args := mock.Called(groupFilters)
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
	values := make([]MonitoredValue, 2)
	value := MonitoredValue{
		MonitorTitle:   "Toto",
		MonitorValue:   "OK",
		MonitorLastRun: time.Now(),
		MonitorStatus:  "OK",
		ServerID:       "158",
		GroupID:        "154",
	}
	value2 := MonitoredValue{
		MonitorTitle:   "Albert",
		MonitorValue:   "Not OK",
		MonitorLastRun: time.Now(),
		MonitorStatus:  "Not OK",
		ServerID:       "158",
		GroupID:        "154",
	}
	values[0] = value
	values[1] = value2
	m := MonitoredValues{
		values,
	}
	api.On("GetResources", mock.Anything).Return(&m, nil)
	ch := make(chan prometheus.Metric)

	groups := make([]GroupFilter, 1)
	gr := GroupFilter{GroupName: "toto"}
	groups[0] = gr
	config = Config{
		ServerURL: "https://hello.com",
		APIKey:    "1234",
		Groups:    groups,
	}
	go func() {
		collector.Collect(ch)
		close(ch)
	}()
	readOne := false
	for m := range ch {
		got := readMetric(m)
		assert.Equal(t, dto.MetricType_UNTYPED, got.metricType)
		if readOne {
			assert.Equal(t, float64(0), got.value)
		} else {
			assert.Equal(t, float64(1), got.value)
		}
		readOne = true
	}
	assert.True(t, readOne)
}

func TestCollector_Collect_NoMetric(t *testing.T) {
	api := MockPAExternalAPI{}
	collector := NewCollector(&api)
	api.On("GetResources", mock.Anything).Return(&MonitoredValues{}, errors.New("error in get resources"))
	ch := make(chan prometheus.Metric)

	groups := make([]GroupFilter, 1)
	gr := GroupFilter{GroupName: "toto"}
	groups[0] = gr
	config = Config{
		ServerURL: "https://hello.com",
		APIKey:    "1234",
		Groups:    groups,
	}
	go func() {
		collector.Collect(ch)
		close(ch)
	}()
	readOne := false
	for m := range ch {
		assert.Contains(t, m.Desc().String(), "poweradmin_error")
		readOne = true
	}
	assert.True(t, readOne)
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

func TestLoadConfig(t *testing.T) {
	s := loadStatuses("status_mapping.yml")
	assert.Equal(t, float64(1), s.Statuses["ok"])
	assert.Equal(t, float64(0), s.Default)
}
