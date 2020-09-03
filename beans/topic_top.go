package beans

import (
	"fmt"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/gonsole/tools/gopsutil/cpu"
	"github.com/lixianmin/gonsole/tools/gopsutil/mem"
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
	TopicTopBody
}

type TopicTopBody struct {
	UpTime       string `json:"uptime"`
	CpuUsage     string `json:"cpu"`
	SysMemory    string `json:"sys"`
	TotalMemory  string `json:"total"`
	NumGoroutine int    `json:"numGoroutine"`
	PauseTotalNs uint64 `json:"pauseTotalNs"`
	NumGC        uint32 `json:"numGC"`
}

func NewTopicTop() *TopicTop {
	var bean = &TopicTop{}
	bean.TopicTopBody = *NewTopicTopData()

	return bean
}

func NewTopicTopData() *TopicTopBody {
	var body = &TopicTopBody{}
	body.NumGoroutine = runtime.NumGoroutine()

	// cpu
	cpuPercent, err := cpu.Percent(0, true)
	if err == nil {
		var list = make([]string, 0, len(cpuPercent))
		for i := range cpuPercent {
			list = append(list, fmt.Sprintf("%.1f%%", cpuPercent[i]))
		}

		var text = "[" + strings.Join(list, ", ") + "]"
		body.CpuUsage = text
	}

	// memory
	var memStats = runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	body.SysMemory = convert.ToHuman(memStats.Sys)
	body.PauseTotalNs = memStats.PauseTotalNs
	body.NumGC = memStats.NumGC

	vm, err := mem.VirtualMemory()
	if err == nil {
		body.TotalMemory = convert.ToHuman(vm.Total)
	}

	body.UpTime = tools.FormatDuration(time.Now().Sub(startProcessTime))
	return body
}
