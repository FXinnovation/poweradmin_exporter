package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	configPath    = kingpin.Flag("config.dir", "Exporter configuration folder.").Default("config").String()
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9575").String()
	metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
)

// Config collection of config files
type Config struct {
	ServerURL     string        `yaml:"server"`
	APIKey        string        `yaml:"api_key"`
	SkipTLSVerify bool          `yaml:"skip_tls_verify"`
	Database      string        `yaml:"database"`
	Groups        []GroupFilter `yaml:"group"`
	StatusMapping StatusConfig  `yaml:"statusMapping"`
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

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading the config: %v", err)
	}

	skipTLS := false
	if config.SkipTLSVerify {
		skipTLS = true
	}
	powerAdminClient, err := NewPAExternalAPIClient(config.APIKey, config.ServerURL, skipTLS)
	if err != nil {
		log.Fatal(err)
	}

	collector := NewCollector(powerAdminClient, config)
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

func loadConfig(configDir string) (Config, error) {

	configuration := Config{}
	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		return configuration, err
	}
	configData := make([]byte, 0)
	for _, file := range files {
		log.Infof("Parsing %s/%s config file", configDir, file.Name())
		configThisData, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", configDir, file.Name()))
		if err != nil {
			return configuration, err
		}
		configData = append(configData, configThisData...)
	}

	errYAML := yaml.Unmarshal([]byte(configData), &configuration)
	if errYAML != nil {
		return configuration, errYAML
	}
	return configuration, nil
}
