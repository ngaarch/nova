//go:build windows

package sysinfo

// DiskSpace holds storage capacity statistics.
type DiskSpace struct {
	TotalBytes uint64  `json:"total_bytes"`
	FreeBytes  uint64  `json:"free_bytes"`
	UsedBytes  uint64  `json:"used_bytes"`
	UsedPct    float64 `json:"used_pct"`
	Available  bool    `json:"available"`
}

func getDiskSpace(path string) DiskSpace {
	return DiskSpace{Available: false}
}
