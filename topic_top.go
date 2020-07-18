package gonsole

import (
	"fmt"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/convert"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"runtime"
	"strings"
	"time"
)

/********************************************************************
created:    2020-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type TopicTop struct {
	BasicResponse
	UpTime       string `json:"uptime"`
	NumGoroutine int    `json:"numGoroutine"`
	CpuUsage     string `json:"cpu"`
	Sys          string `json:"sys"`
	TotalMemory  string `json:"total"`
	PauseTotalNs uint64 `json:"pauseTotalNs"`
	NumGC        uint32 `json:"numGC"`
}

func newTopicTop() *TopicTop {
	var bean = &TopicTop{}
	bean.Operation = "top"
	bean.Timestamp = tools.GetTimestamp()
	bean.NumGoroutine = runtime.NumGoroutine()

	// cpu
	cpuPercent, err := cpu.Percent(0, true)
	if err == nil {
		var list = make([]string, 0, len(cpuPercent))
		for i := range cpuPercent {
			list = append(list, fmt.Sprintf("%.1f%%", cpuPercent[i]))
		}

		var text = "[" + strings.Join(list, ", ") + "]"
		bean.CpuUsage = text
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

	bean.UpTime = tools.FormatDuration(time.Now().Sub(startProcessTime))
	return bean
}
