package cnc

// Host stats — small Pi-side health readout the operator sees in the
// /machine topbar. CPU temp, load avg, memory + disk %, uptime. Reads
// /proc + /sys files directly (Linux-only path; macOS / Windows return
// zeros). Cheap enough that the 10-second poll the frontend does has
// no measurable impact even on a Pi 4.

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// HostStats is a snapshot of Pi health metrics. Zero values are
// returned when a particular file isn't readable (e.g. CPU temp on a
// VM that doesn't have a thermal zone). Callers should render "—" for
// any field they get 0 on.
type HostStats struct {
	TempC       float64 `json:"temp_c,omitempty"`
	Load1m      float64 `json:"load_1m,omitempty"`
	MemUsedPct  float64 `json:"mem_used_pct,omitempty"`
	DiskUsedPct float64 `json:"disk_used_pct,omitempty"`
	UptimeS     int64   `json:"uptime_s,omitempty"`
	Cores       int     `json:"cores,omitempty"`
}

// ReadHostStats reads the live values. Best-effort — every field is
// optional. Order is fixed-cost: ~5 small file reads + one statfs.
func ReadHostStats() HostStats {
	out := HostStats{Cores: countCores()}
	out.TempC = readThermalC()
	out.Load1m = readLoadAvg1m()
	out.MemUsedPct = readMemUsedPct()
	out.DiskUsedPct = readDiskUsedPct("/")
	out.UptimeS = readUptimeS()
	return out
}

func readThermalC() float64 {
	// Pi 4 exposes SoC temp at zone 0. Value is millidegrees C.
	// Some boards expose multiple zones; zone 0 is the SoC on the Pi.
	b, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	if err != nil {
		return 0
	}
	return v / 1000.0
}

func readLoadAvg1m() float64 {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(b))
	if len(fields) < 1 {
		return 0
	}
	v, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0
	}
	return v
}

func readMemUsedPct() float64 {
	// MemTotal - MemAvailable is the "used" half of the typical
	// `free -m` rendering. MemAvailable accounts for reclaimable
	// page cache so we don't false-alarm on a Pi running Chromium.
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0
	}
	defer f.Close()
	var total, avail float64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "MemTotal:"):
			total = parseKB(line)
		case strings.HasPrefix(line, "MemAvailable:"):
			avail = parseKB(line)
		}
		if total > 0 && avail > 0 {
			break
		}
	}
	if total <= 0 {
		return 0
	}
	used := total - avail
	if used < 0 {
		used = 0
	}
	return (used / total) * 100.0
}

// parseKB pulls the integer field out of a `Key:    1234 kB` line.
func parseKB(line string) float64 {
	idx := strings.IndexByte(line, ':')
	if idx < 0 {
		return 0
	}
	rest := strings.TrimSpace(line[idx+1:])
	// Strip trailing " kB" if present.
	if i := strings.IndexByte(rest, ' '); i >= 0 {
		rest = rest[:i]
	}
	v, err := strconv.ParseFloat(rest, 64)
	if err != nil {
		return 0
	}
	return v
}

func readDiskUsedPct(path string) float64 {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0
	}
	if st.Blocks == 0 {
		return 0
	}
	used := st.Blocks - st.Bavail
	return float64(used) / float64(st.Blocks) * 100.0
}

func readUptimeS() int64 {
	b, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(b))
	if len(fields) < 1 {
		return 0
	}
	v, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0
	}
	return int64(v)
}

func countCores() int {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return 0
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "processor") {
			n++
		}
	}
	return n
}
