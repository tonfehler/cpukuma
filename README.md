# cpukuma

A lightweight Go agent that monitors CPU and RAM usage and pushes alerts to [Uptime Kuma](https://github.com/louislam/uptime-kuma) when thresholds are exceeded.

Works on **macOS**, **Linux**, and **Windows**.

---

## How it works

On each run, cpukuma:

1. Reads current CPU load (1-second sample) and RAM usage percentage
2. Compares each value against its configured threshold
3. Sends an HTTP push notification to Uptime Kuma if a threshold is exceeded

Run it on a schedule (cron, Task Scheduler, systemd timer) at whatever interval you need.

---

## Configuration

On first run, cpukuma creates `config/config.json` with defaults:

```json
{
  "cpu_push_url": "<Uptime Kuma Push URL>",
  "cpualert": 50.0,
  "mem_push_url": "<Uptime Kuma Push URL>",
  "memalert": 50.0
}
```

| Key | Description |
|---|---|
| `cpu_push_url` | Uptime Kuma push URL for CPU alerts |
| `cpualert` | CPU usage threshold in percent (e.g. `80.0`) |
| `mem_push_url` | Uptime Kuma push URL for RAM alerts |
| `memalert` | RAM usage threshold in percent (e.g. `90.0`) |

Replace the placeholder URLs with the push URLs from your Uptime Kuma monitor (Monitor → Edit → Copy Push URL).

---

## Building

From the repository root:
```bash
go build -o cpukuma src/main.go
```

---

## Running

```bash
./cpukuma
```

To run every minute on Linux/macOS, add a crontab entry:

```
* * * * * /path/to/cpukuma && ./cpukuma
```

---

## Requirements

- Go 1.21+
- An [Uptime Kuma](https://github.com/louislam/uptime-kuma) instance with push monitors configured
- github.com/shirou/gopsutil/v3/mem
