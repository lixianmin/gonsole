package logger

import (
	"fmt"
	"os"
	"strings"
)

/********************************************************************
created:    2020-04-25
author:     lixianmin

Copyright (C) - All Rights Reserved
 *********************************************************************/

type ILogger interface {
	Info(first interface{}, args ...interface{})
	Warn(first interface{}, args ...interface{})
	Error(first interface{}, args ...interface{})
}

type ConsoleLogger struct {
}

func (l *ConsoleLogger) Info(first interface{}, args ...interface{}) {
	var text = formatLog(first, args...)
	writeMessage(os.Stdout, text)
}

func (l *ConsoleLogger) Warn(first interface{}, args ...interface{}) {
	var text = formatLog(first, args...)
	writeMessage(os.Stderr, text)
}

func (l *ConsoleLogger) Error(first interface{}, args ...interface{}) {
	var text = formatLog(first, args...)
	writeMessage(os.Stderr, text)
}

func writeMessage(fout *os.File, text string) {
	_, _ = fout.WriteString(text)
	_, _ = fout.WriteString("\n")
}

func formatLog(first interface{}, args ...interface{}) string {
	var msg string
	switch first.(type) {
	case string:
		msg = first.(string)
		if len(args) == 0 {
			return msg
		}
		if strings.Contains(msg, "%") && !strings.Contains(msg, "%%") {
			//format string
		} else {
			//do not contain format char
			msg += strings.Repeat(" %v", len(args))
		}
	default:
		msg = fmt.Sprint(first)
		if len(args) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(args))
	}
	return fmt.Sprintf(msg, args...)
}