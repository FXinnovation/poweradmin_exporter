package poweradmin

import (
	"encoding/xml"
	"fmt"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	paServerString    = "%s?KEY=%s"
	monitorInfoSuffix = "&API=GET_MONITOR_INFO&XML=1&CID=%s"
)

type MonitorInfos struct {
	Infos []MonitorInfo `xml:"monitor"`
}

// MonitorInfo return of the GET_MONITOR_INFO call
type MonitorInfo struct {
	Id      string `xml:"id,attr"`
	Status  string `xml:"status,attr"`
	Title   string `xml:"title,attr"`
	LastRun paTime `xml:"lastRun,attr"`
}

type paTime struct {
	time.Time
}

func (c *paTime) UnmarshalXMLAttr(attr xml.Attr) error {
	const shortForm = "02-01-2006 15:04:05"

	parse, err := time.Parse(shortForm, attr.Value)
	if err != nil {
		return err
	}
	*c = paTime{parse}
	return nil
}

type powerAdminConnection struct {
	ApiKey       string
	ServerUrl    string
	ApiUrlString string
	Client       *http.Client
}

func createPAClient(apiKey string, serverUrl string) powerAdminConnection {
	return powerAdminConnection{
		ApiKey:    apiKey,
		ServerUrl: serverUrl,
		Client:    &http.Client{},
	}
}

// MonitorInfoClient is a client to power admin
type MonitorInfoClient powerAdminConnection

// NewMonitorInfoClient creates a client for monitor info calls
func NewMonitorInfoClient(apiKey string, serverUrl string) *MonitorInfoClient {
	pa := createPAClient(apiKey, serverUrl)
	mi := MonitorInfoClient(pa)
	mi.ApiUrlString = fmt.Sprintf(paServerString, serverUrl, apiKey) + monitorInfoSuffix
	return &mi
}

func (pa powerAdminConnection) sendRequest(req *http.Request) ([]byte, error) {
	resp, err := pa.Client.Do(req)
	if err != nil {
		log.Errorf("Error sending request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading response body: %v", err)
		return nil, err
	}
	return data, err
}

func (m MonitorInfoClient) GetMonitorInfos(cid string) (*MonitorInfos, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(m.ApiUrlString, cid), nil)
	if err != nil {
		log.Errorf("Error building GetMonitorInfo request: %v", err)
		return nil, err
	}

	resp, err := powerAdminConnection(m).sendRequest(req)
	if err != nil {
		log.Errorf("Error querying %s: %s", req.RequestURI, err.Error())
		return nil, err
	}
	monitors := &MonitorInfos{}
	err = xml.Unmarshal(resp, monitors)
	if err != nil {
		log.Errorf("Error unmarshalling response %s", err)
		return nil, err
	}
	return monitors, nil
}
