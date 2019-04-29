package poweradmin

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	paServerString    = "%s?KEY=%s"
	monitorInfoSuffix = "&API=GET_MONITOR_INFO&XML=1&CID=%s"
	groupListSuffix   = "&API=GET_GROUP_LIST&XML=1"
	serverListSuffix  = "&API=GET_SERVER_LIST&XML=1&GID=%s"
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

// GroupList return of the GET_GROUP_LIST call
type GroupList struct {
	Groups []Group `xml:"group"`
}

// Group return of the GET_GROUP_LIST call
type Group struct {
	ID       string `xml:"id,attr"`
	Name     string `xml:"name,attr"`
	Path     string `xml:"path,attr"`
	ParentID string `xml:"parentID,attr"`
}

// ServerList return of the GET_Server_LIST call
type ServerList struct {
	Servers []Server `xml:"server"`
}

// Server return of the GET_Server_LIST call
type Server struct {
	ID      string `xml:"id,attr"`
	Name    string `xml:"name,attr"`
	Alias   string `xml:"alias,attr"`
	Status  string `xml:"status,attr"`
	GroupID string `xml:"groupID,attr"`
	Group   string `xml:"group,attr"`
}

// MonitoredValues the values retrieved
type MonitoredValues struct {
	Values []MonitoredValue
}

// MonitoredValue one value with its attributes
type MonitoredValue struct {
	MonitorTitle   string
	MonitorValue   string
	MonitorStatus  string
	MonitorLastRun time.Time
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
	APIKey         string
	ServerURL      string
	MonitorInfoURL string
	GroupListURL   string
	ServerListURL  string
	Client         *http.Client
}

func createPAClient(apiKey string, serverURL string) PAExternalAPIClient {
	return PAExternalAPIClient{
		APIKey:    apiKey,
		ServerURL: serverURL,
		Client:    &http.Client{},
	}
}

// NewPAExternalAPIClient creates a client for monitor info calls
func NewPAExternalAPIClient(apiKey string, serverURL string) (*PAExternalAPIClient, error) {
	if apiKey == "" {
		return nil, errors.New("API Key cannot be empty")
	}
	pa := createPAClient(apiKey, serverURL)
	paURL := fmt.Sprintf(paServerString, serverURL, apiKey)
	pa.MonitorInfoURL = paURL + monitorInfoSuffix
	pa.GroupListURL = paURL + groupListSuffix
	pa.ServerListURL = paURL + serverListSuffix
	return &pa, nil
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

func (client PAExternalAPIClient) getResponse(requestURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Errorf("Error building request: %v", err)
		return nil, err
	}

	resp, err := client.sendRequest(req)
	if err != nil {
		log.Errorf("Error querying %s: %s", req.RequestURI, err.Error())
		return nil, err
	}
	return resp, err
}

// GetMonitorInfos returns monitorinfos for a cid
func (client PAExternalAPIClient) GetMonitorInfos(cid string) (*MonitorInfos, error) {
	resp, err := client.getResponse(fmt.Sprintf(client.MonitorInfoURL, cid))
	if err != nil {
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

// GetGroupList returns all groups
func (client PAExternalAPIClient) GetGroupList() (*GroupList, error) {
	resp, err := client.getResponse(client.GroupListURL)
	if err != nil {
		return nil, err
	}
	groups := &GroupList{}
	err = xml.Unmarshal(resp, groups)
	if err != nil {
		log.Errorf("Error unmarshalling response %s", err)
		return nil, err
	}
	return groups, nil
}

// GetServerList returns all groups
func (client PAExternalAPIClient) GetServerList(gid string) (*ServerList, error) {
	resp, err := client.getResponse(fmt.Sprintf(client.ServerListURL, gid))
	if err != nil {
		return nil, err
	}
	servers := &ServerList{}
	err = xml.Unmarshal(resp, servers)
	if err != nil {
		log.Errorf("Error unmarshalling response %s", err)
		return nil, err
	}
	return servers, nil
}

func (client PAExternalAPIClient) GetResources(gname string) (*MonitoredValues, error) {
	groups, err := client.GetGroupList()
	if err != nil {
		return nil, err
	}
	for _, group := range groups.Groups {
		if group.Name == gname {
			servers, err := client.GetServerList(group.ID)
			if err != nil {
				return nil, err
			}
			for _, server := range servers.Servers {
				values, err := client.GetMonitorInfos(server.ID)
				if err != nil {
					return nil, err
				}
				metrics := MonitoredValues{}
				metrics.Values = make([]MonitoredValue, len(values.Infos))
			}
			break
		}
	}
	return nil, errors.New("No group found")
}
