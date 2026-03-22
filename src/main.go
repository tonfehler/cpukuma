package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

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

		fmt.Println("Neue config erstellt:", configPath)
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
		fmt.Println("Error reading CPU load:", err)
		return 0.0
	}

	load := percents[0]
	fmt.Printf("CPU Load: %.2f%%\n", load)
	return load
}

/*
Function getRAMUsage
Reads the system RAM usage percentage, cross-platform.
*/
func getRAMUsage() float64 {
	v, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Error reading RAM usage:", err)
		return 0.0
	}

	fmt.Printf("RAM Usage: %.2f%%\n", v.UsedPercent)
	return v.UsedPercent
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("Fehler beim Laden der Config:", err)
		os.Exit(1)
	}

	fmt.Printf("Config geladen: URL=%s\n", cfg.CPUPushURL)

	if getCPULoad() >= cfg.CpuAlertThreshold {
		resp, err := http.Get(cfg.CPUPushURL)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Fehler beim Lesen der Antwort:", err)
			return
		}
		fmt.Println(string(body))
	}

	if getRAMUsage() >= cfg.MemAlertThreshold {
		resp, err := http.Get(cfg.MemPushURL)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Fehler beim Lesen der Antwort:", err)
			return
		}
		fmt.Println(string(body))
	}
}
