package services

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type SystemMetrics struct {
	CPU     CPUMetrics     `json:"cpu"`
	Memory  MemoryMetrics  `json:"memory"`
	Disk    DiskMetrics    `json:"disk"`
	Network NetworkMetrics `json:"network"`
	Updated time.Time      `json:"updated"`
}

type CPUMetrics struct {
	UsagePercent float64   `json:"usagePercent"`
	Cores        int       `json:"cores"`
	PerCore      []float64 `json:"perCore"`
}

type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"usedPercent"`
}

type DiskMetrics struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

type NetworkMetrics struct {
	BytesSent   uint64 `json:"bytesSent"`
	BytesRecv   uint64 `json:"bytesRecv"`
	PacketsSent uint64 `json:"packetsSent"`
	PacketsRecv uint64 `json:"packetsRecv"`
}

type MonitoringService struct {
	metrics     SystemMetrics
	mutex       sync.RWMutex
	stopChan    chan bool
	interval    time.Duration
	lastNetStat net.IOCountersStat
}

func NewMonitoringService() *MonitoringService {
	return &MonitoringService{
		interval: 5 * time.Second,
		stopChan: make(chan bool),
	}
}

func (s *MonitoringService) Start() {
	go s.collectMetrics()
}

func (s *MonitoringService) Stop() {
	s.stopChan <- true
}

func (s *MonitoringService) GetMetrics() SystemMetrics {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.metrics
}

func (s *MonitoringService) collectMetrics() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Collect initial metrics
	s.updateMetrics()

	for {
		select {
		case <-ticker.C:
			s.updateMetrics()
		case <-s.stopChan:
			return
		}
	}
}

func (s *MonitoringService) updateMetrics() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// CPU metrics
	cpuPercent, _ := cpu.Percent(time.Second, false)
	cpuPerCore, _ := cpu.Percent(time.Second, true)
	cpuCount, _ := cpu.Counts(true)
	cpuUsage := 0.0
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	s.metrics.CPU = CPUMetrics{
		UsagePercent: cpuUsage,
		Cores:        cpuCount,
		PerCore:      cpuPerCore,
	}

	// Memory metrics
	memInfo, _ := mem.VirtualMemory()
	if memInfo != nil {
		s.metrics.Memory = MemoryMetrics{
			Total:       memInfo.Total,
			Used:        memInfo.Used,
			Available:   memInfo.Available,
			UsedPercent: memInfo.UsedPercent,
		}
	}

	// Disk metrics (root partition)
	diskInfo, _ := disk.Usage("/")
	if diskInfo != nil {
		s.metrics.Disk = DiskMetrics{
			Total:       diskInfo.Total,
			Used:        diskInfo.Used,
			Free:        diskInfo.Free,
			UsedPercent: diskInfo.UsedPercent,
		}
	}

	// Network metrics
	netStats, _ := net.IOCounters(false)
	if len(netStats) > 0 {
		s.metrics.Network = NetworkMetrics{
			BytesSent:   netStats[0].BytesSent,
			BytesRecv:   netStats[0].BytesRecv,
			PacketsSent: netStats[0].PacketsSent,
			PacketsRecv: netStats[0].PacketsRecv,
		}
	}

	s.metrics.Updated = time.Now()
}
