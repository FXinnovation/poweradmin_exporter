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

// MonitorInfos return of the GET_MONITOR_INFO call
type MonitorInfos struct {
	Infos []MonitorInfo `xml:"monitor"`
}

// MonitorInfo return of the GET_MONITOR_INFO call
type MonitorInfo struct {
	ID      string `xml:"id,attr"`
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

// PAExternalAPIClient client for PowerAdmin External API struct
type PAExternalAPIClient struct {
	APIKey       string
	ServerURL    string
	APIURLString string
	Client       *http.Client
}

func createPAClient(apiKey string, serverURL string) PAExternalAPIClient {
	return PAExternalAPIClient{
		APIKey:    apiKey,
		ServerURL: serverURL,
		Client:    &http.Client{},
	}
}

// NewPAExternalAPIClient creates a client for monitor info calls
func NewPAExternalAPIClient(apiKey string, serverURL string) *PAExternalAPIClient {
	pa := createPAClient(apiKey, serverURL)
	pa.APIURLString = fmt.Sprintf(paServerString, serverURL, apiKey) + monitorInfoSuffix
	return &pa
}

func (client PAExternalAPIClient) sendRequest(req *http.Request) ([]byte, error) {
	resp, err := client.Client.Do(req)
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

// GetMonitorInfos returns monitorinfos for a cid
func (client PAExternalAPIClient) GetMonitorInfos(cid string) (*MonitorInfos, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(client.APIURLString, cid), nil)
	if err != nil {
		log.Errorf("Error building GetMonitorInfo request: %v", err)
		return nil, err
	}

	resp, err := client.sendRequest(req)
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
