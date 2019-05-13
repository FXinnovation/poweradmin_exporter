package main

import (
	"crypto/tls"
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
	MonitorID      string
	MonitorTitle   string
	MonitorValue   string
	MonitorStatus  string
	MonitorLastRun time.Time
	ServerID       string
	ServerName     string
	GroupID        string
	GroupName      string
	GroupPath      string
}

type paTime struct {
	time.Time
}

func (c *paTime) UnmarshalXMLAttr(attr xml.Attr) error {
	const shortForm = "02-01-2006 15:04:05"

	parse, err := time.Parse(shortForm, attr.Value)
	if err != nil {
		log.Errorf("Error parsing %s", attr.Value)
		parse = time.Now()
	}
	*c = paTime{parse}
	return nil
}

// PAExternalAPI interface
type PAExternalAPI interface {
	GetMonitorInfos(cid string) (*MonitorInfos, error)
	GetGroupList() (*GroupList, error)
	GetServerList(gid string) (*ServerList, error)
	GetResources(groupFilters []GroupFilter) (*MonitoredValues, error)
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

func createPAClient(apiKey string, serverURL string, skipTLSVerify bool) PAExternalAPIClient {

	transCfg := &http.Transport{}
	if skipTLSVerify {
		transCfg = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		}
	}
	return PAExternalAPIClient{
		APIKey:    apiKey,
		ServerURL: serverURL,
		Client:    &http.Client{Transport: transCfg},
	}
}

// NewPAExternalAPIClient creates a client for monitor info calls
func NewPAExternalAPIClient(apiKey string, serverURL string, skipTLSVerify bool) (*PAExternalAPIClient, error) {
	if apiKey == "" {
		return nil, errors.New("API Key cannot be empty")
	}

	pa := createPAClient(apiKey, serverURL, skipTLSVerify)
	paURL := fmt.Sprintf(paServerString, serverURL, apiKey)
	pa.MonitorInfoURL = paURL + monitorInfoSuffix
	pa.GroupListURL = paURL + groupListSuffix
	pa.ServerListURL = paURL + serverListSuffix
	return &pa, nil
}

func sendRequest(req *http.Request, client *http.Client) ([]byte, error) {
	resp, err := client.Do(req)
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

func getResponse(requestURL string, client *http.Client) ([]byte, error) {
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Errorf("Error building request: %v", err)
		return nil, err
	}
	resp, err := sendRequest(req, client)
	if err != nil {
		log.Errorf("Error querying %s: %s", req.RequestURI, err.Error())
		return nil, err
	}
	return resp, err
}

// GetMonitorInfos returns monitorinfos for a cid
func (client *PAExternalAPIClient) GetMonitorInfos(cid string) (*MonitorInfos, error) {
	resp, err := getResponse(fmt.Sprintf(client.MonitorInfoURL, cid), client.Client)
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
func (client *PAExternalAPIClient) GetGroupList() (*GroupList, error) {
	resp, err := getResponse(client.GroupListURL, client.Client)
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
func (client *PAExternalAPIClient) GetServerList(gid string) (*ServerList, error) {
	resp, err := getResponse(fmt.Sprintf(client.ServerListURL, gid), client.Client)
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

// GetResources get the monitor values for a group name
func (client *PAExternalAPIClient) GetResources(groupFilters []GroupFilter) (*MonitoredValues, error) {
	groups, err := client.GetGroupList()
	if err != nil {
		return nil, err
	}
	groupSet := make(map[string]Group, len(groups.Groups))

	// populate once a set so that we don't search the groupNames slice several times for contains
	for _, group := range groups.Groups {
		groupSet[group.Path] = group
	}
	metrics := MonitoredValues{}
	metrics.Values = make([]MonitoredValue, 0)
	for _, filter := range groupFilters {
		group, groupExists := groupSet[filter.GroupPath]
		if groupExists {
			servers, err := client.GetServerList(group.ID)
			if err != nil {
				return nil, err
			}
			filteredServers := filterServers(servers.Servers, filter.Servers)
			for _, server := range filteredServers {
				values, err := client.GetMonitorInfos(server.ID)
				if err != nil {
					return nil, err
				}
				metricTitles := make(map[string]int)
				for _, metric := range values.Infos {
					// metric title is not unique
					if _, titleExists := metricTitles[metric.Title]; titleExists {
						metricTitles[metric.Title]++
						log.Warnf("Duplicate monitor %s for server %s!", metric.Title, server.Name)
					} else {
						metricTitles[metric.Title] = 1

						newMetric := MonitoredValue{
							GroupID:        group.ID,
							GroupName:      group.Name,
							GroupPath:      group.Path,
							ServerID:       server.ID,
							ServerName:     server.Name,
							MonitorValue:   metric.Status,
							MonitorStatus:  metric.Status,
							MonitorTitle:   metric.Title,
							MonitorLastRun: metric.LastRun.Time,
							MonitorID:      metric.ID,
						}
						metrics.Values = append(metrics.Values, newMetric)
					}
				}
			}
		} else {
			log.Warnf("group named %s not found", filter)
		}
	}
	return &metrics, nil
}

func filterServers(servers []Server, names []string) []Server {
	if len(names) == 0 { // if no server names in filter, then it's like no filter
		return servers
	}
	namesSet := make(map[string]struct{}, len(names))
	for _, name := range names {
		namesSet[name] = struct{}{}
	}
	newServers := make([]Server, 0)
	for _, server := range servers {
		if _, exists := namesSet[server.Name]; exists {
			newServers = append(newServers, server)
		}
	}
	return newServers

}
