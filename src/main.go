package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

/*
Struct Config
The following struct defines key-value pairs for our config file
*/
type Config struct {
	CPUPushURL        string  `json:"cpu_push_url"`
	CpuAlertThreshold float64 `json:"cpualert"`
	MemPushURL        string  `json:"mem_push_url"`
	MemAlertThreshold float64 `json:"memualert"`
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
Reads the System CPU Load with OS functions
*/
func getCPULoad() float64 {
	// Output looks like: "{ 1.23 2.34 3.45 }"
	out, err := exec.Command("sysctl", "-n", "vm.loadavg").Output()
	if err != nil {
		fmt.Println("Error reading CPU load:", err)
		return 0.0
	}

	s := strings.Trim(strings.TrimSpace(string(out)), "{} ")
	fields := strings.Fields(s)
	if len(fields) < 1 {
		fmt.Println("Unexpected sysctl output:", string(out))
		return 0.0
	}

	load, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		fmt.Println("Error parsing CPU load:", err)
		return 0.0
	}

	fmt.Printf("CPU Load (1m avg): %.2f\n", load)
	return load
}

/*
Function getRAMUsage
Reads the System RAM usage as a percentage using vm_stat and sysctl
*/
func getRAMUsage() float64 {
	// Get total physical memory in bytes
	totalOut, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		fmt.Println("Error reading total RAM:", err)
		return 0.0
	}
	totalBytes, err := strconv.ParseInt(strings.TrimSpace(string(totalOut)), 10, 64)
	if err != nil {
		fmt.Println("Error parsing total RAM:", err)
		return 0.0
	}

	// Get page statistics from vm_stat
	vmOut, err := exec.Command("vm_stat").Output()
	if err != nil {
		fmt.Println("Error reading vm_stat:", err)
		return 0.0
	}

	pageSize := int64(4096)
	var active, inactive, wired int64

	for line := range strings.SplitSeq(string(vmOut), "\n") {
		var val int64
		if strings.HasPrefix(line, "Mach Virtual Memory Statistics") {
			// Extract page size: "page size of 16384 bytes"
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "size" && i+2 < len(parts) {
					if ps, err := strconv.ParseInt(parts[i+2], 10, 64); err == nil {
						pageSize = ps
					}
				}
			}
		}
		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			continue
		}
		valStr := strings.TrimRight(strings.TrimSpace(fields[1]), ".")
		val, err = strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			continue
		}
		switch strings.TrimSpace(fields[0]) {
		case "Pages active":
			active = val
		case "Pages inactive":
			inactive = val
		case "Pages wired down":
			wired = val
		}
	}

	usedBytes := (active + inactive + wired) * pageSize
	usage := float64(usedBytes) / float64(totalBytes) * 100.0

	fmt.Printf("RAM Usage: %.2f%%\n", usage)
	return usage
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("Fehler beim Laden der Config:", err)
		os.Exit(1)
	}

	fmt.Printf("Config geladen: URL=%s, Interval=%ds\n", cfg.CPUPushURL)

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
