package collector

type Metrics struct {
	Hostname           string          `json:"hostname,omitempty"`
	IpAddress          string          `json:"ip_address,omitempty"`
	Os                 string          `json:"os,omitempty"`
	CpuUsage           *float64        `json:"cpu_usage,omitempty"`
	CpuLoad1           *float64        `json:"cpu_load_1,omitempty"`
	CpuLoad5           *float64        `json:"cpu_load_5,omitempty"`
	CpuLoad15          *float64        `json:"cpu_load_15,omitempty"`
	MemoryTotal        *uint64         `json:"memory_total,omitempty"`
	MemoryUsed         *uint64         `json:"memory_used,omitempty"`
	MemoryAvailable    *uint64         `json:"memory_available,omitempty"`
	MemoryUsagePercent *float64        `json:"memory_usage_percent,omitempty"`
	SwapTotal          *uint64         `json:"swap_total,omitempty"`
	SwapUsed           *uint64         `json:"swap_used,omitempty"`
	SwapUsagePercent   *float64        `json:"swap_usage_percent,omitempty"`
	DiskMetrics        []DiskMetric    `json:"disk_metrics,omitempty"`
	NetworkMetrics     []NetworkMetric `json:"network_metrics,omitempty"`
	CollectedAt        string          `json:"collected_at"`
}

type DiskMetric struct {
	MountPoint     string  `json:"mount_point"`
	TotalBytes     uint64  `json:"total_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
}

type NetworkMetric struct {
	InterfaceName string `json:"interface_name"`
	RxBytes       uint64 `json:"rx_bytes"`
	TxBytes       uint64 `json:"tx_bytes"`
	RxPackets     uint64 `json:"rx_packets"`
	TxPackets     uint64 `json:"tx_packets"`
	RxErrors      uint64 `json:"rx_errors"`
	TxErrors      uint64 `json:"tx_errors"`
}
