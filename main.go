package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

const (
	interfaceConfigFileName = "config.yml"
	filterConfigFileName    = "filter.yml"
	statusMappingFileName   = "status_mapping.yml"
)

var (
	configPath    = kingpin.Flag("config.dir", "Exporter configuration folder.").Default("config").String()
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9575").String()
	metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	config        Config
)

// Config collection of config files
type Config struct {
	Interface     InterfaceConfig
	Filter        FilterConfig
	StatusMapping StatusConfig
}

// InterfaceConfig configures the interfaces to use
type InterfaceConfig struct {
	ServerURL     string `yaml:"server"`
	APIKey        string `yaml:"api_key"`
	SkipTLSVerify bool   `yaml:"skip_tls_verify"`
}

// FilterConfig configures the filter to apply
type FilterConfig struct {
	Groups []GroupFilter `yaml:"group"`
}

// GroupFilter group selection
type GroupFilter struct {
	GroupPath string   `yaml:"path"`
	Servers   []string `yaml:"servers"`
}

// StatusConfig configure the status values to be sent
type StatusConfig struct {
	Statuses map[string]float64 `yaml:"values"`
	Default  float64            `yaml:"default"`
}

func init() {
	prometheus.MustRegister(version.NewCollector("poweradmin_exporter"))
}

func main() {
	kingpin.Version(version.Print("poweradmin_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Info("Starting exporter ", version.Info())
	log.Info("Build context ", version.BuildContext())

	config := Config{
		Interface:     InterfaceConfig{},
		Filter:        FilterConfig{},
		StatusMapping: StatusConfig{},
	}
	err := loadConfig(*configPath, statusMappingFileName, config.StatusMapping)
	if err != nil {
		log.Fatalf("Error loading the config: %v", err)
	}
	err = loadConfig(*configPath, interfaceConfigFileName, config.Interface)
	if err != nil {
		log.Fatalf("Error loading the config: %v", err)
	}
	err = loadConfig(*configPath, filterConfigFileName, config.Filter)
	if err != nil {
		log.Fatalf("Error loading the config: %v", err)
	}

	skipTLS := false
	if config.Interface.SkipTLSVerify {
		skipTLS = true
	}
	powerAdminClient, err := NewPAExternalAPIClient(config.Interface.APIKey, config.Interface.ServerURL, skipTLS)
	if err != nil {
		log.Fatal(err)
	}

	collector := NewCollector(powerAdminClient)
	prometheus.MustRegister(collector)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>PowerAdmin Exporter</title></head>
			<body>
			<h1>PowerAdmin Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
		if err != nil {
			log.Fatal(fmt.Errorf("Error writing index page content: %s", err))
		}
	})

	log.Infof("Beginning to serve on address %v", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func loadConfig(configDir string, configName string, config interface{}) error {

	// Load the config from the file
	configData, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", configDir, configName))

	if err != nil {
		return err
	}

	errYAML := yaml.Unmarshal([]byte(configData), &config)
	if errYAML != nil {
		return errYAML
	}

	return nil
}
