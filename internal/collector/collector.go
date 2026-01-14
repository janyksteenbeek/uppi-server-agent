package collector

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// Collect gathers all system metrics and returns them
func Collect() (*Metrics, error) {
	metrics := &Metrics{
		CollectedAt: time.Now().UTC().Format(time.RFC3339),
	}

	collectHostInfo(metrics)
	collectCPUMetrics(metrics)
	collectMemoryMetrics(metrics)
	collectDiskMetrics(metrics)
	collectNetworkMetrics(metrics)

	return metrics, nil
}

func collectHostInfo(metrics *Metrics) {
	// Hostname
	if hostname, err := os.Hostname(); err == nil {
		metrics.Hostname = hostname
	}

	// OS info
	if hostInfo, err := host.Info(); err == nil {
		metrics.Os = fmt.Sprintf("%s %s", hostInfo.Platform, hostInfo.PlatformVersion)
	}

	// Main IP address
	metrics.IpAddress = getMainIPAddress()
}

func collectCPUMetrics(metrics *Metrics) {
	// CPU usage percentage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics.CpuUsage = &cpuPercent[0]
	}

	// Load averages
	loadAvg, err := load.Avg()
	if err == nil {
		metrics.CpuLoad1 = &loadAvg.Load1
		metrics.CpuLoad5 = &loadAvg.Load5
		metrics.CpuLoad15 = &loadAvg.Load15
	}
}

func collectMemoryMetrics(metrics *Metrics) {
	// Virtual memory
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		metrics.MemoryTotal = &memInfo.Total
		metrics.MemoryUsed = &memInfo.Used
		metrics.MemoryAvailable = &memInfo.Available
		metrics.MemoryUsagePercent = &memInfo.UsedPercent
	}

	// Swap memory
	swapInfo, err := mem.SwapMemory()
	if err == nil {
		metrics.SwapTotal = &swapInfo.Total
		metrics.SwapUsed = &swapInfo.Used
		metrics.SwapUsagePercent = &swapInfo.UsedPercent
	}
}

func collectDiskMetrics(metrics *Metrics) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		metrics.DiskMetrics = append(metrics.DiskMetrics, DiskMetric{
			MountPoint:     partition.Mountpoint,
			TotalBytes:     usage.Total,
			UsedBytes:      usage.Used,
			AvailableBytes: usage.Free,
			UsagePercent:   usage.UsedPercent,
		})
	}
}

func collectNetworkMetrics(metrics *Metrics) {
	netStats, err := net.IOCounters(true)
	if err != nil {
		return
	}

	for _, stat := range netStats {
		// Skip loopback interface
		if stat.Name == "lo" {
			continue
		}

		metrics.NetworkMetrics = append(metrics.NetworkMetrics, NetworkMetric{
			InterfaceName: stat.Name,
			RxBytes:       stat.BytesRecv,
			TxBytes:       stat.BytesSent,
			RxPackets:     stat.PacketsRecv,
			TxPackets:     stat.PacketsSent,
			RxErrors:      stat.Errin,
			TxErrors:      stat.Errout,
		})
	}
}

func getMainIPAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range interfaces {
		// Skip loopback and interfaces with no addresses
		if iface.Name == "lo" || len(iface.Addrs) == 0 {
			continue
		}

		for _, addr := range iface.Addrs {
			ip := addr.Addr

			// Remove CIDR notation if present
			if idx := strings.Index(ip, "/"); idx != -1 {
				ip = ip[:idx]
			}

			// Skip IPv6 and loopback addresses
			if strings.Contains(ip, ":") || ip == "127.0.0.1" {
				continue
			}

			// Return the first valid IPv4 address
			if ip != "" {
				return ip
			}
		}
	}

	return ""
}
