package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const (
	groupListString = `
<?xml version="1.0"?>
<groups>
    <group name="Servers/Devices" path="Servers/Devices" id="0" parentID="-1"/>
    <group name="Central" path="Servers/Devices^Live^Central" id="2" parentID="647"/>
    <group name="FX" path="Servers/Devices^Live^FX" id="193" parentID="647"/>
</groups>
`
	monitorString = `
<monitors>
<monitor id="8937" status="OK" depends_on="" title="Ping FXMACHINE1" lastRun="10-04-2019 13:18:28" nextRun="10-04-2019 13:19:28" errText="[Last response: 1 ms] " errActionIDs="570" fixedActionIDs="570" inErrSeconds="0" />
</monitors>
`
	emptyMonitorList = `
<monitors/>
`
	monitorUnmarshallErrorString = `
<monitors>
<monitor id="8937" status="OK" depends_on="" title="Ping" lastRun="NoTime 13:18:28" nextRun="10-04-2019 13:19:28" errText="[Last response: 1 ms] " errActionIDs="570" fixedActionIDs="570" inErrSeconds="0" />
</monitors>
`
	monitorStringNotAnXML = `
?Not an xml
`
	serverListString = `
<?xml version="1.0"?>
<servers>
<server name="FXH1" group="Servers/Devices^Live^FX Hosting^Physical^FX^Prod" id="568" groupID="193" alias="FXH1" status="ok"/>
<server name="FXH2" group="Servers/Devices^Live^FX Hosting^Physical^FX^Prod" id="709" groupID="193" alias="FXH2" status="ok"/>
<server name="FXH3" group="Servers/Devices^Live^FX Hosting^Physical^FX^Prod" id="710" groupID="193" alias="FXH3" status="ok"/>
</servers>
`
)

func TestNewPAExternalAPIClient(t *testing.T) {
	monitor, _ := NewPAExternalAPIClient("1234key", "https://serverpa", false)
	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.MonitorInfoURL)
	assert.Contains(t, monitor.MonitorInfoURL, "GET_MONITOR_INFO")
	assert.NotNil(t, monitor.Client)
	assert.Equal(t, reflect.TypeOf(&http.Client{}), reflect.TypeOf(monitor.Client))
}

func TestNewPAExternalAPIClient_GetMonitorInfos(t *testing.T) {
	// returns a monitor xml
	monitorHandler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(monitorString))
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(monitorHandler))
	defer ts.Close()
	monitor, _ := NewPAExternalAPIClient("1234key", ts.URL, false)
	monitors, err := monitor.GetMonitorInfos("ALL")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(monitors.Infos))
	assert.Equal(t, "8937", monitors.Infos[0].ID)
	assert.Equal(t, "OK", monitors.Infos[0].Status)
	assert.Equal(t, "Ping FXMACHINE1", monitors.Infos[0].Title)
	assert.Equal(t, 2019, monitors.Infos[0].LastRun.Year())
}

func TestNewPAExternalAPIClient_GetMonitorInfos_NoResponse(t *testing.T) {
	monitor, _ := NewPAExternalAPIClient("1234key", "https://nourl.com", false)
	_, err := monitor.GetMonitorInfos("ALL")
	assert.NotNil(t, err)
}

func TestNewPAExternalAPIClient_GetMonitorInfos_BadUrl(t *testing.T) {
	monitor, _ := NewPAExternalAPIClient("1234key", "::?s::s&t::::oto\x20--notanurl.com", false)
	_, err := monitor.GetMonitorInfos("ALL")
	assert.NotNil(t, err)
}

func TestNewPAExternalAPIClient_GetMonitorInfos_UnmarshalError(t *testing.T) {
	// returns a monitor xml
	monitorHandler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(monitorUnmarshallErrorString))
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(monitorHandler))
	defer ts.Close()
	monitor, _ := NewPAExternalAPIClient("1234key", ts.URL, false)
	_, err := monitor.GetMonitorInfos("ALL")
	assert.NotNil(t, err)
}

func TestNewPAExternalAPIClient_GetMonitorInfos_TimeParsingError(t *testing.T) {
	// returns a monitor xml
	monitorHandler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(monitorStringNotAnXML))
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(monitorHandler))
	defer ts.Close()
	monitor, _ := NewPAExternalAPIClient("1234key", ts.URL, false)
	_, err := monitor.GetMonitorInfos("ALL")
	assert.NotNil(t, err)
}

func TestNewPAExternalAPIClient_GetGroupList(t *testing.T) {
	// returns a grouplist xml
	groupHandler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(groupListString))
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(groupHandler))
	defer ts.Close()
	client, _ := NewPAExternalAPIClient("1234key", ts.URL, false)
	groups, err := client.GetGroupList()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(groups.Groups))
	assert.Equal(t, "2", groups.Groups[1].ID)
	assert.Equal(t, "Central", groups.Groups[1].Name)
	assert.Equal(t, "Servers/Devices^Live^Central", groups.Groups[1].Path)
	assert.Equal(t, "647", groups.Groups[1].ParentID)
}

func TestNewPAExternalAPIClient_GetServerList(t *testing.T) {
	// returns a serverlist xml
	serverHandler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(serverListString))
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(serverHandler))
	defer ts.Close()
	client, _ := NewPAExternalAPIClient("1234key", ts.URL, false)
	servers, err := client.GetServerList("193")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(servers.Servers))
	assert.Equal(t, "568", servers.Servers[0].ID)
	assert.Equal(t, "FXH1", servers.Servers[0].Name)
	assert.Equal(t, "Servers/Devices^Live^FX Hosting^Physical^FX^Prod", servers.Servers[0].Group)
	assert.Equal(t, "193", servers.Servers[0].GroupID)
	assert.Equal(t, "ok", servers.Servers[0].Status)
	assert.Equal(t, "FXH1", servers.Servers[0].Alias)
}

func TestNewPAExternalAPIClient_NoApiKey(t *testing.T) {
	client, err := NewPAExternalAPIClient("", "server", false)
	assert.Nil(t, client)
	assert.NotNil(t, err)
}

func TestPAExternalAPIClient_GetResources(t *testing.T) {
	resourcesHandler := func(w http.ResponseWriter, r *http.Request) {
		apiParam := r.URL.Query()["API"][0]
		if apiParam == "GET_GROUP_LIST" {
			_, _ = w.Write([]byte(groupListString))
		} else if apiParam == "GET_SERVER_LIST" {
			_, _ = w.Write([]byte(serverListString))
		} else if apiParam == "GET_MONITOR_INFO" {
			if r.URL.Query()["CID"][0] == "568" {
				_, _ = w.Write([]byte(monitorString))
			} else {
				_, _ = w.Write([]byte(emptyMonitorList))
			}
		}
	}
	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(resourcesHandler))
	defer ts.Close()
	client, _ := NewPAExternalAPIClient("1234key", ts.URL, false)
	metrics, err := client.GetResources([]GroupFilter{{GroupName: "FX"}})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(metrics.Values))
	assert.Equal(t, "OK", metrics.Values[0].MonitorValue)
}

func TestPAExternalAPIClient_GetResources_NoGroups(t *testing.T) {
	resourcesHandler := func(w http.ResponseWriter, r *http.Request) {
		apiParam := r.URL.Query()["API"][0]
		if apiParam == "GET_GROUP_LIST" {
			_, _ = w.Write([]byte(groupListString))
		} else if apiParam == "GET_SERVER_LIST" {
			_, _ = w.Write([]byte(serverListString))
		} else if apiParam == "GET_MONITOR_INFO" {
			if r.URL.Query()["CID"][0] == "568" {
				_, _ = w.Write([]byte(monitorString))
			} else {
				_, _ = w.Write([]byte(emptyMonitorList))
			}
		}
	}
	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(resourcesHandler))
	defer ts.Close()
	client, _ := NewPAExternalAPIClient("1234key", ts.URL, false)
	metrics, err := client.GetResources([]GroupFilter{{GroupName: "NOFX"}})
	assert.Equal(t, 0, len(metrics.Values))
	assert.Nil(t, err)
}

func TestPAExternalAPIClient_GetResources_NoServers(t *testing.T) {
	resourcesHandler := func(w http.ResponseWriter, r *http.Request) {
		apiParam := r.URL.Query()["API"][0]
		if apiParam == "GET_GROUP_LIST" {
			_, _ = w.Write([]byte(groupListString))
		} else if apiParam == "GET_SERVER_LIST" {
			_, _ = w.Write([]byte("<servers/>"))
		} else if apiParam == "GET_MONITOR_INFO" {
			if r.URL.Query()["CID"][0] == "568" {
				_, _ = w.Write([]byte(monitorString))
			} else {
				_, _ = w.Write([]byte(emptyMonitorList))
			}
		}
	}
	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(resourcesHandler))
	defer ts.Close()
	client, _ := NewPAExternalAPIClient("1234key", ts.URL, false)
	metrics, err := client.GetResources([]GroupFilter{{GroupName: "FX"}})
	assert.NotNil(t, metrics)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(metrics.Values))
}

func TestPAExternalAPIClient_GetResources_FilterServer(t *testing.T) {
	resourcesHandler := func(w http.ResponseWriter, r *http.Request) {
		apiParam := r.URL.Query()["API"][0]
		if apiParam == "GET_GROUP_LIST" {
			_, _ = w.Write([]byte(groupListString))
		} else if apiParam == "GET_SERVER_LIST" {
			_, _ = w.Write([]byte(serverListString))
		} else if apiParam == "GET_MONITOR_INFO" {
			if r.URL.Query()["CID"][0] == "568" {
				_, _ = w.Write([]byte(monitorString))
			} else {
				_, _ = w.Write([]byte(emptyMonitorList))
			}
		}
	}
	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(resourcesHandler))
	defer ts.Close()
	client, _ := NewPAExternalAPIClient("1234key", ts.URL, false)
	metrics, err := client.GetResources([]GroupFilter{{GroupName: "FX", Servers: []string{"FXH1"}}})
	assert.Equal(t, 1, len(metrics.Values))
	assert.Nil(t, err)

	metrics, err = client.GetResources([]GroupFilter{{GroupName: "FX", Servers: []string{"FXH2"}}})
	assert.Equal(t, 0, len(metrics.Values))
	assert.Nil(t, err)
}
