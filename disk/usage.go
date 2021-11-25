package disk

import "github.com/shirou/gopsutil/disk"

func GetDiskUsage(path string) *disk.UsageStat {
	usage, err := disk.Usage(path)
	if err != nil {
		return nil
	}
	return usage
}
