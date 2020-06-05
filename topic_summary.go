package gonsole

import (
	"fmt"
	"github.com/lixianmin/gocore/convert"
	"github.com/lixianmin/gonsole/tools"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"runtime"
	"time"
)

type TopicSummary struct {
	BasicResponse
	UpTime       string `json:"uptime"`
	NumGoroutine int    `json:"numGoroutine"`
	CpuUsage     string `json:"cpu"`
	Sys          string `json:"sys"`
	TotalMemory  string `json:"total"`
	PauseTotalNs uint64 `json:"pauseTotalNs"`
	NumGC        uint32 `json:"numGC"`
}

func newTopicSummary() *TopicSummary {
	var bean = &TopicSummary{}
	bean.Operation = "summary"
	bean.Timestamp = tools.GetTimestamp()
	bean.NumGoroutine = runtime.NumGoroutine()

	// cpu
	cpuPercent, err := cpu.Percent(0, true)
	if err == nil {
		var text = "["
		for idx, percent := range cpuPercent {
			if idx == 0 {
				text += fmt.Sprintf("%.1f%%", percent)
			} else {
				text += fmt.Sprintf(", %.1f%%", percent)
			}
		}

		bean.CpuUsage = text + "]"
	}

	// memory
	var memStats = runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	bean.Sys = convert.ToHuman(memStats.Sys)
	bean.PauseTotalNs = memStats.PauseTotalNs
	bean.NumGC = memStats.NumGC

	vm, err := mem.VirtualMemory()
	if err == nil {
		bean.TotalMemory = convert.ToHuman(vm.Total)
	}

	bean.UpTime = time.Now().Sub(startProcessTime).String()
	return bean
}
