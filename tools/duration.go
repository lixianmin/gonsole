package tools

import (
	"time"
)

/********************************************************************
created:    2020-07-17
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func FormatDuration(d time.Duration) string {
	var buffer = make([]byte, 0, 8)

	const oneDay = time.Hour * 24
	var days time.Duration
	if d > oneDay {
		days = d / oneDay
		d = d - days*oneDay

		Itoa(&buffer, int(days), -1)
		buffer = append(buffer, 'd')
	}

	var hours time.Duration
	if d > time.Hour {
		hours = d / time.Hour
		d = d - hours*time.Hour

		Itoa(&buffer, int(hours), -1)
		buffer = append(buffer, 'h')
	}

	var minutes time.Duration
	if d > time.Minute {
		minutes = d / time.Minute
		d = d - minutes*time.Minute

		Itoa(&buffer, int(minutes), -1)
		buffer = append(buffer, 'm')
	}

	var seconds = d / time.Second
	Itoa(&buffer, int(seconds), -1)
	buffer = append(buffer, 's')

	var text = string(buffer)
	return text
}
