package logger

/********************************************************************
created:    2020-04-25
author:     lixianmin

Copyright (C) - All Rights Reserved
 *********************************************************************/

var defaultLogger ILogger = &ConsoleLogger{}

func Init(log ILogger) {
	if log != nil {
		defaultLogger = log
	}
}

func GetDefaultLogger() ILogger {
	return defaultLogger
}

func Info(first interface{}, args ...interface{}) {
	defaultLogger.Info(first, args...)
}

func Warn(first interface{}, args ...interface{}) {
	defaultLogger.Warn(first, args...)
}

func Error(first interface{}, args ...interface{}) {
	defaultLogger.Error(first, args...)
}
