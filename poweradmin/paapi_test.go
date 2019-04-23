package poweradmin

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewMonitorInfoClient(t *testing.T) {

	monitor := NewMonitorInfoClient("1234key", "https://serverpa")
	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.ApiUrlString)
	assert.Contains(t, monitor.ApiUrlString, "GET_MONITOR_INFO")
	assert.NotNil(t, monitor.Client)
	assert.Equal(t, reflect.TypeOf(&http.Client{}), reflect.TypeOf(monitor.Client))
}

func TestMonitorInfoClient_GetMonitorInfos(t *testing.T) {
	monitorString := `
<monitors>
<monitor id="8937" status="OK" depends_on="" title="Ping FXHTYUL1PMON001" lastRun="10-04-2019 13:18:28" nextRun="10-04-2019 13:19:28" errText="[Last response: 1 ms] " errActionIDs="570" fixedActionIDs="570" inErrSeconds="0" />
</monitors>
`

	// returns a monitor xml
	monitorHandler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(monitorString))
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(monitorHandler))
	defer ts.Close()
	fmt.Print(ts.URL)
	monitor := NewMonitorInfoClient("1234key", ts.URL)
	monitors, err := monitor.GetMonitorInfos("ALL")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(monitors.Infos))
	assert.Equal(t, "8937", monitors.Infos[0].Id)
	assert.Equal(t, "OK", monitors.Infos[0].Status)
	assert.Equal(t, "Ping FXHTYUL1PMON001", monitors.Infos[0].Title)
	assert.Equal(t, 2019, monitors.Infos[0].LastRun.Year())
}

func TestMonitorInfoClient_GetMonitorInfos_NoResponse(t *testing.T) {
	monitor := NewMonitorInfoClient("1234key", "https://nourl.com")
	_, err := monitor.GetMonitorInfos("ALL")
	assert.NotNil(t, err)
}

func TestMonitorInfoClient_GetMonitorInfos_BadUrl(t *testing.T) {
	monitor := NewMonitorInfoClient("1234key", "::?s::s&t::::oto\x20--notanurl.com")
	_, err := monitor.GetMonitorInfos("ALL")
	assert.NotNil(t, err)
}

func TestMonitorInfoClient_GetMonitorInfos_UnmarshalError(t *testing.T) {
	monitorString := `
<monitors>
<monitor id="8937" status="OK" depends_on="" title="Ping FXHTYUL1PMON001" lastRun="NoTime 13:18:28" nextRun="10-04-2019 13:19:28" errText="[Last response: 1 ms] " errActionIDs="570" fixedActionIDs="570" inErrSeconds="0" />
</monitors>
`

	// returns a monitor xml
	monitorHandler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(monitorString))
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(monitorHandler))
	defer ts.Close()
	fmt.Print(ts.URL)
	monitor := NewMonitorInfoClient("1234key", ts.URL)
	_, err := monitor.GetMonitorInfos("ALL")
	assert.NotNil(t, err)
}

func TestMonitorInfoClient_GetMonitorInfos_TimeParsingError(t *testing.T) {
	monitorString := `
?Not an xml
`

	// returns a monitor xml
	monitorHandler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(monitorString))
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(monitorHandler))
	defer ts.Close()
	fmt.Print(ts.URL)
	monitor := NewMonitorInfoClient("1234key", ts.URL)
	_, err := monitor.GetMonitorInfos("ALL")
	assert.NotNil(t, err)
}
