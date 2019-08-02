package main

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/mock"
	"strings"
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

type MockSQLServerConnection struct {
	mock.Mock
}

func (mock *MockSQLServerConnection) connect() error {
	args := mock.Called()
	return args.Error(0)
}

func (mock *MockSQLServerConnection) Close() error {
	args := mock.Called()
	return args.Error(0)
}
func (mock *MockSQLServerConnection) GetConnection() error {
	args := mock.Called()
	return args.Error(0)
}
func (mock *MockSQLServerConnection) GetStatData(serversID []string) ([]StatData, error) {
	args := mock.Called(serversID)
	return args.Get(0).([]StatData), args.Error(1)
}
func (mock *MockSQLServerConnection) GetConfigComputerInfo(alias string) (ConfigComputerInfo, error) {
	args := mock.Called(alias)
	return args.Get(0).(ConfigComputerInfo), args.Error(1)
}
func (mock *MockSQLServerConnection) GetAllServersFor(group string) ([]string, error) {
	args := mock.Called(group)
	return args.Get(0).([]string), args.Error(1)
}
func (mock *MockSQLServerConnection) GetAllServerMetric(serverID int) ([]ServerMetric, error) {
	args := mock.Called(serverID)
	return args.Get(0).([]ServerMetric), args.Error(1)
}

func TestCollector_Collect(t *testing.T) {
	api := MockPAExternalAPI{}

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

	// mock sql server calls
	sql := MockSQLServerConnection{}
	servers := []string{
		"toto",
	}
	sm1 := ServerMetric{
		Name:           "FXSERVER",
		Alias:          "FXSERVER",
		GroupID:        2,
		StatID:         2,
		Value:          4.53,
		Date:           time.Time{},
		OwnerType:      4,
		ItemName:       "W32Time",
		StatName:       "ServiceUp",
		OwningComputer: "9b14ea69-3ec6-4e3f-ae27",
		CompID:         2,
		StatType:       11,
		ItemAlias:      "Windows Time",
		Unit:           16,
		UnitStr:        "",
	}
	compInfo := ConfigComputerInfo{
		CompID:  2,
		Name:    "FXSERVER",
		Alias:   "FXSERVER",
		GroupID: 2,
	}
	sql.On("GetConnection").Return(nil)
	sql.On("GetConfigComputerInfo", mock.Anything).Return(compInfo, nil)
	sql.On("GetAllServersFor", mock.Anything).Return(servers, nil)
	sql.On("GetAllServerMetric", mock.Anything).Return([]ServerMetric{
		sm1,
	}, nil)
	ch := make(chan prometheus.Metric)

	groups := make([]GroupFilter, 1)
	gr := GroupFilter{GroupPath: "toto"}
	groups[0] = gr
	statuses := make(map[string]float64, 1)
	statuses["ok"] = 1
	config := Config{
		ServerURL: "https://hello.com",
		APIKey:    "1234",
		Groups:    groups,
		StatusMapping: StatusConfig{
			Statuses: statuses,
			Default:  0,
		},
	}
	collector := NewCollector(&api, &sql, config)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()
	nbReadMetrics := 0
	for m := range ch {
		got := readMetric(m)

		if got.metricType != dto.MetricType_UNTYPED {
			t.Errorf("Wrong metric type: got %v, want %v", got.metricType, dto.MetricType_UNTYPED)
		}
		if nbReadMetrics == 0 {
			if got.value != float64(1) {
				t.Errorf("Wrong value: got %v, want %v", got.value, float64(1))
			}
		} else if nbReadMetrics == 1 {
			if got.value != float64(0) {
				t.Errorf("Wrong value: got %v, want %v", got.value, float64(0))
			}
		} else {
			if got.value != 4.53 {
				t.Errorf("Wrong value: got %v, want %v", got.value, 4.53)
			}
		}
		nbReadMetrics++
	}

	if nbReadMetrics != 3 {
		t.Errorf("Wrong number of metrics received: got %v, want %v", nbReadMetrics, 3)
	}
}

func TestCollector_Collect_NoMetric(t *testing.T) {
	api := MockPAExternalAPI{}
	api.On("GetResources", mock.Anything).Return(&MonitoredValues{}, errors.New("error in get resources"))

	// mock sql server calls
	sql := MockSQLServerConnection{}
	servers := []string{
		"toto",
	}
	compInfo := ConfigComputerInfo{
		CompID:  2,
		Name:    "FXSERVER",
		Alias:   "FXSERVER",
		GroupID: 2,
	}
	sql.On("GetConnection").Return(nil)
	sql.On("GetConfigComputerInfo", mock.Anything).Return(compInfo)
	sql.On("GetAllServersFor", mock.Anything).Return(servers, nil)
	sql.On("GetAllServerMetric", mock.Anything).Return([]ServerMetric{}, nil)
	ch := make(chan prometheus.Metric)

	groups := make([]GroupFilter, 1)
	gr := GroupFilter{GroupPath: "toto"}
	groups[0] = gr
	config := Config{
		ServerURL: "https://hello.com",
		APIKey:    "1234",
		Groups:    groups,
	}
	collector := NewCollector(&api, &sql, config)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()
	readOne := false
	for m := range ch {
		if !strings.Contains(m.Desc().String(), "poweradmin_error") {
			t.Errorf("Description doesn't contain value %v: got %v", "poweradmin_error", m.Desc().String())
		}
		readOne = true
	}

	if readOne != true {
		t.Errorf("Wrong value: got %v, want %v", readOne, true)
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

func Test_getFormattedMetricName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal 1", args{"power on"}, "power_on"},
		{"backslash", args{"\\\\Windows\\Folder input pages/sec"}, "__windows_folder_input_pages_per_sec"},
		{"trailing underscore", args{"power on you    "}, "power_on_you"},
		{"trailing underscore 2", args{"end with ____"}, "end_with"},
		{"starting digit", args{"1234I love cats"}, "_1234i_love_cats"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFormattedMetricName(tt.args.name); got != tt.want {
				t.Errorf("getFormattedMetricName() = %v, want %v", got, tt.want)
			}
		})
	}
}
