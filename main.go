package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var version = "undefined"

type Check struct {
	Name           string        `yaml:"name"`
	URL            string        `yaml:"url"`
	ExpectedStatus int           `yaml:"expected_status"`
	Timeout        time.Duration `yaml:"timeout"`
}

type Config struct {
	ListenAddress string             `yaml:"listen_address"`
	Checks        map[string][]Check `yaml:"checks"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.ListenAddress == "" {
		cfg.ListenAddress = ":9001"
	}

	if cfg.Checks == nil {
		cfg.Checks = make(map[string][]Check)
	}

	return &cfg, nil
}

func checkURL(ch Check) error {
	timeout := ch.Timeout
	if timeout == 0 {
		timeout = 1 * time.Second
	}

	client := http.Client{Timeout: timeout}
	resp, err := client.Get(ch.URL)
	if err != nil {
		return fmt.Errorf("check %s failed: %v", ch.Name, err)
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body) // drain body

	if resp.StatusCode != ch.ExpectedStatus {
		return fmt.Errorf("check %s: unexpected status %d (expected %d)",
			ch.Name, resp.StatusCode, ch.ExpectedStatus)
	}
	return nil
}

func main() {
	configFlag := flag.String("config", "config.yml", "path to config file")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return
	}

	configPath := *configFlag
	if configPath == "config.yml" && flag.NArg() > 0 {
		configPath = flag.Arg(0)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	totalChecks := 0
	for name, list := range cfg.Checks {
		log.Printf("Loaded %d checks for %s", len(list), name)
		totalChecks += len(list)
	}
	log.Printf("Loaded total %d checks from %s", totalChecks, configPath)

	// aggregate handler for all checks
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" && r.URL.Path != "/healthz/" {
			http.NotFound(w, r)
			return
		}
		for groupName, checks := range cfg.Checks {
			for _, ch := range checks {
				if err := checkURL(ch); err != nil {
					log.Printf("Health check failed (%s/%s): %v", groupName, ch.Name, err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("FAIL\n"))
					return
				}
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	// handler for named groups: /healthz/<name>
	http.HandleFunc("/healthz/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/healthz/")
		if name == "" {
			// no name provided, treat as aggregate
			http.Redirect(w, r, "/healthz", http.StatusSeeOther)
			return
		}

		checks, ok := cfg.Checks[name]
		if !ok {
			http.NotFound(w, r)
			return
		}

		for _, ch := range checks {
			if err := checkURL(ch); err != nil {
				log.Printf("Health check failed (%s/%s): %v", name, ch.Name, err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("FAIL\n"))
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	log.Printf("Health aggregator listening on %s", cfg.ListenAddress)
	log.Fatal(http.ListenAndServe(cfg.ListenAddress, nil))
}
