package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var logger *log.Logger

/*
funtion initLogger()
This Method handels creation of the log file and folder
*/
func initLogger() error {
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("logs-Verzeichnis konnte nicht erstellt werden: %w", err)
	}

	f, err := os.OpenFile("logs/log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("log-Datei konnte nicht geöffnet werden: %w", err)
	}

	logger = log.New(f, "", 0)
	return nil
}

/*
function logf
Support-Function to handle writing logs.
*/
func logf(format string, args ...any) {
	ts := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	logger.Printf("[%s] %s", ts, msg)
}

/*
Struct Config
The following struct defines key-value pairs for our config file
*/
type Config struct {
	CPUPushURL        string  `json:"cpu_push_url"`
	CpuAlertThreshold float64 `json:"cpualert"`
	MemPushURL        string  `json:"mem_push_url"`
	MemAlertThreshold float64 `json:"memalert"`
}

/*
function loadConfig() This Method loads and if needed creates the configuration file in config/
*/
func loadConfig() (*Config, error) {
	const configPath = "config/config.json"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultCfg := Config{
			CPUPushURL:        "<Uptime Kuma Push URL>",
			CpuAlertThreshold: 50.0,
			MemPushURL:        "<Uptime Kuma Push URL>",
			MemAlertThreshold: 50.0,
		}

		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return nil, fmt.Errorf("config-Verzeichnis konnte nicht erstellt werden: %w", err)
		}

		data, err := json.MarshalIndent(defaultCfg, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("config konnte nicht serialisiert werden: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return nil, fmt.Errorf("config konnte nicht geschrieben werden: %w", err)
		}

		logf("Neue config erstellt: %s", configPath)
		return &defaultCfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config konnte nicht gelesen werden: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config konnte nicht geparst werden: %w", err)
	}

	return &cfg, nil
}

/*
Function getCPULoad
Reads the system CPU usage percentage (1-second sample), cross-platform.
*/
func getCPULoad() float64 {
	percents, err := cpu.Percent(time.Second, false)
	if err != nil || len(percents) == 0 {
		logf("Error reading CPU load: %v", err)
		return 0.0
	}

	load := percents[0]
	logf("CPU Load: %.2f%%", load)
	return load
}

/*
Function getRAMUsage
Reads the system RAM usage percentage, cross-platform.
*/
func getRAMUsage() float64 {
	v, err := mem.VirtualMemory()
	if err != nil {
		logf("Error reading RAM usage: %v", err)
		return 0.0
	}

	logf("RAM Usage: %.2f%%", v.UsedPercent)
	return v.UsedPercent
}

func main() {
	if err := initLogger(); err != nil {
		fmt.Fprintln(os.Stderr, "Logger konnte nicht initialisiert werden:", err)
		os.Exit(1)
	}

	cfg, err := loadConfig()
	if err != nil {
		logf("Fehler beim Laden der Config: %v", err)
		os.Exit(1)
	}

	logf("Config geladen: URL=%s", cfg.CPUPushURL)

	if getCPULoad() >= cfg.CpuAlertThreshold {
		resp, err := http.Get(cfg.CPUPushURL)
		if err != nil {
			logf("%v", err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logf("Fehler beim Lesen der Antwort: %v", err)
			return
		}
		logf("%s", string(body))
	}

	if getRAMUsage() >= cfg.MemAlertThreshold {
		resp, err := http.Get(cfg.MemPushURL)
		if err != nil {
			logf("%v", err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logf("Fehler beim Lesen der Antwort: %v", err)
			return
		}
		logf("%s", string(body))
	}
}
