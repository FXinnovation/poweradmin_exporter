package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
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
	if monitor == nil {
		t.Errorf("Monitor shouldn't be nil: got %v", monitor)
	}
	if len(monitor.MonitorInfoURL) == 0 {
		t.Errorf("MonitorInfoURL shouldn't be empty: got %v", monitor.MonitorInfoURL)
	}
	if !strings.Contains(monitor.MonitorInfoURL, "GET_MONITOR_INFO") {
		t.Errorf("MonitorInfoURL doesn't contain value %v: got %v", "GET_MONITOR_INFO", monitor.MonitorInfoURL)
	}
	if monitor.Client == nil {
		t.Errorf("Monitor.Client shouldn't be nil: got %v", monitor.Client)
	}
	if reflect.TypeOf(&http.Client{}) != reflect.TypeOf(monitor.Client) {
		t.Errorf("Monitor.Client should be %v", reflect.TypeOf(&http.Client{}))
	}
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
	if err != nil {
		t.Errorf("Error should be nil: got %v", err)
	}
	if len(monitors.Infos) != 1 {
		t.Errorf("Wrong size for monitors.Infos: got %v, want %v", len(monitors.Infos), 1)
	}

	if monitors.Infos[0].ID != "8937" {
		t.Errorf("Wrong value for monitors.Infos[0].ID: got %v, want %v", monitors.Infos[0].ID, "8937")
	}

	if monitors.Infos[0].Status != "OK" {
		t.Errorf("Wrong value for monitors.Infos[0].Status: got %v, want %v", monitors.Infos[0].Status, "OK")
	}

	if monitors.Infos[0].Title != "Ping FXMACHINE1" {
		t.Errorf("Wrong value for monitors.Infos[0].Title: got %v, want %v", monitors.Infos[0].Title, "Ping FXMACHINE1")
	}

	if monitors.Infos[0].LastRun.Year() != 2019 {
		t.Errorf("Wrong value for monitors.Infos[0].LastRun.Year(): got %v, want %v", monitors.Infos[0].LastRun.Year(), 2019)
	}
}

func TestNewPAExternalAPIClient_GetMonitorInfos_NoResponse(t *testing.T) {
	monitor, _ := NewPAExternalAPIClient("1234key", "https://nourl.com", false)
	_, err := monitor.GetMonitorInfos("ALL")
	if err == nil {
		t.Errorf("Error shouldn't be nil: got %v", err)
	}
}

func TestNewPAExternalAPIClient_GetMonitorInfos_BadUrl(t *testing.T) {
	monitor, _ := NewPAExternalAPIClient("1234key", "::?s::s&t::::oto\x20--notanurl.com", false)
	_, err := monitor.GetMonitorInfos("ALL")
	if err == nil {
		t.Errorf("Error shouldn't be nil: got %v", err)
	}
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
	if err == nil {
		t.Errorf("Error shouldn't be nil: got %v", err)
	}
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
	if err == nil {
		t.Errorf("Error shouldn't be nil: got %v", err)
	}
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
	if err != nil {
		t.Errorf("Error should be nil: got %v", err)
	}

	if len(groups.Groups) != 3 {
		t.Errorf("Wrong size for groups.Groups: got %v, want %v", len(groups.Groups), 3)
	}

	if groups.Groups[1].ID != "2" {
		t.Errorf("Wrong value for groups.Groups[1].ID: got %v, want %v", groups.Groups[1].ID, "2")
	}

	if groups.Groups[1].Name != "Central" {
		t.Errorf("Wrong value for groups.Groups[1].Name: got %v, want %v", groups.Groups[1].Name, "Central")
	}

	if groups.Groups[1].Path != "Servers/Devices^Live^Central" {
		t.Errorf("Wrong value for groups.Groups[1].Path: got %v, want %v", groups.Groups[1].Path, "Servers/Devices^Live^Central")
	}

	if groups.Groups[1].ParentID != "647" {
		t.Errorf("Wrong value for groups.Groups[1].ParentID: got %v, want %v", groups.Groups[1].ParentID, "647")
	}
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
	if err != nil {
		t.Errorf("Error should be nil: got %v", err)
	}

	if len(servers.Servers) != 3 {
		t.Errorf("Wrong size for servers.Servers: got %v, want %v", len(servers.Servers), 3)
	}

	if servers.Servers[0].ID != "568" {
		t.Errorf("Wrong value for servers.Servers[0].ID: got %v, want %v",servers.Servers[0].ID, "568")
	}

	if servers.Servers[0].Name != "FXH1" {
		t.Errorf("Wrong value for servers.Servers[0].Name: got %v, want %v", servers.Servers[0].Name, "FXH1")
	}

	if servers.Servers[0].Group != "Servers/Devices^Live^FX Hosting^Physical^FX^Prod" {
		t.Errorf("Wrong value for servers.Servers[0].Path: got %v, want %v", servers.Servers[0].Group, "Servers/Devices^Live^FX Hosting^Physical^FX^Prod")
	}

	if servers.Servers[0].GroupID != "193" {
		t.Errorf("Wrong value for servers.Servers[0].ParentID: got %v, want %v", servers.Servers[0].GroupID, "193")
	}

	if servers.Servers[0].Status != "ok" {
		t.Errorf("Wrong value for servers.Servers[0].Status: got %v, want %v", servers.Servers[0].Status, "ok")
	}

	if servers.Servers[0].Alias != "FXH1" {
		t.Errorf("Wrong value for servers.Servers[0].Alias: got %v, want %v", servers.Servers[0].Alias, "FXH1")
	}
}

func TestNewPAExternalAPIClient_NoApiKey(t *testing.T) {
	client, err := NewPAExternalAPIClient("", "server", false)

	if client != nil {
		t.Errorf("Client should be nil: got %v", client)
	}
	if err == nil {
		t.Errorf("Error shouldn't be nil: got %v", err)
	}
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
	if err != nil {
		t.Errorf("Error should be nil: got %v", err)
	}
	if len(metrics.Values) != 1 {
		t.Errorf("Wrong size for metrics.Values: got %v, want %v", len(metrics.Values), 1)
	}
	if metrics.Values[0].MonitorValue != "OK" {
		t.Errorf("Wrong value for metrics.Values[0].MonitorValue: got %v, want %v", metrics.Values[0].MonitorValue, "OK")
	}
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
	if len(metrics.Values) != 0 {
		t.Errorf("Wrong size for metrics.Values: got %v, want %v", len(metrics.Values), 0)
	}
	if err != nil {
		t.Errorf("Error should be nil: got %v", err)
	}
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
	if metrics == nil {
		t.Errorf("Metrics shouldn't be nil: got %v", metrics)
	}
	if err != nil {
		t.Errorf("Error should be nil: got %v", err)
	}

	if len(metrics.Values) != 0 {
		t.Errorf("Wrong size for metrics.Values: got %v, want %v", len(metrics.Values), 0)
	}
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
	if len(metrics.Values) != 1 {
		t.Errorf("Wrong size: got %v, want %v", len(metrics.Values), 1)
	}
	if err != nil {
		t.Errorf("Error should be nil: got %v", err)
	}

	metrics, err = client.GetResources([]GroupFilter{{GroupName: "FX", Servers: []string{"FXH2"}}})
	if len(metrics.Values) != 0 {
		t.Errorf("Wrong size: got %v, want %v", len(metrics.Values), 0)
	}
	if err != nil {
		t.Errorf("Error should be nil: got %v", err)
	}
}
