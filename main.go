package main

import (
	"fmt"
	"github.com/prometheus/common/log"
	"io/ioutil"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"net/http"

	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	configFile    = kingpin.Flag("config.file", "Exporter configuration file.").Default("config.yml").String()
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9575").String()
	metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	config        Config
)

// Config of the exporter
type Config struct {
	ServerURL     string        `yaml:"server"`
	APIKey        string        `yaml:"api_key"`
	Groups        []GroupFilter `yaml:"group"`
	SkipTLSVerify bool          `yaml:"skip_tls_verify"`
}

// GroupFilter group selection
type GroupFilter struct {
	GroupPath string   `yaml:"path"`
	Servers   []string `yaml:"servers"`
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

	config, err := loadConfig(*configFile)
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

func loadConfig(configFile string) (Config, error) {
	config := Config{}

	// Load the config from the file
	configData, err := ioutil.ReadFile(configFile)

	if err != nil {
		return config, err
	}

	errYAML := yaml.Unmarshal([]byte(configData), &config)
	if errYAML != nil {
		return config, errYAML
	}

	return config, nil
}
