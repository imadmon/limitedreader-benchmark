package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type MonitorValueType string

var (
	SyntheticRXBytes atomic.Uint64

	RX               MonitorValueType = "RX"
	SyntheticRX      MonitorValueType = "SyntheticRX"
	TotalSyntheticRX MonitorValueType = "TotalSyntheticRX"
	CPU              MonitorValueType = "CPU"
	RAM              MonitorValueType = "RAM"
)

type monitorResult struct {
	rxDelta          uint64
	syntheticRXDelta uint64
	totalSyntheticRX uint64
	cpuPercent       float64
	ramMB            float64
}

func monitorLoop(ctx context.Context, resultsC chan []monitorResult) {
	SyntheticRXBytes.Store(0) // reset for monitor
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	currRx, err := getRX()
	if err != nil {
		fmt.Printf("Error reading RX bytes: %v\n", err)
	}

	var prevRx uint64 = currRx
	var prevSyntheticRx uint64
	results := make([]monitorResult, 0)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Monitor loop stopped.")
			resultsC <- results
			return
		case <-ticker.C:
			currRx, err := getRX()
			if err != nil {
				fmt.Printf("Error reading RX bytes: %v\n", err)
				continue
			}
			rxDelta := currRx - prevRx
			prevRx = currRx

			currSyntheticRx := SyntheticRXBytes.Load()
			syntheticRxDelta := currSyntheticRx - prevSyntheticRx
			prevSyntheticRx = currSyntheticRx

			cpuPercent, err := cpu.Percent(0, false)
			if err != nil || len(cpuPercent) == 0 {
				fmt.Printf("Error reading CPU usage: %v\n", err)
				continue
			}

			vmStat, err := mem.VirtualMemory()
			if err != nil {
				fmt.Printf("Error reading memory usage: %v\n", err)
				continue
			}
			ramMB := float64(vmStat.Used) / 1024.0 / 1024.0

			results = append(results, monitorResult{
				rxDelta:          rxDelta,
				syntheticRXDelta: syntheticRxDelta,
				totalSyntheticRX: currSyntheticRx,
				cpuPercent:       cpuPercent[0],
				ramMB:            ramMB,
			})
			fmt.Printf("RX: %d bytes |CPU: %.2f%% | RAM: %.2fMB | SyntheticRX: %d bytes | TotalSyntheticRX: %d bytes\n",
				rxDelta, cpuPercent[0], ramMB, syntheticRxDelta, currSyntheticRx)
		}
	}
}

func getRX() (uint64, error) {
	ioCounters, err := net.IOCounters(false)
	if err != nil || len(ioCounters) == 0 {
		return 0, fmt.Errorf("error getting ioCounters: %v", err)
	}
	if len(ioCounters) == 0 {
		return 0, fmt.Errorf("error got 0 ioCounters")
	}
	return ioCounters[0].BytesRecv, nil
}
