package intern

import (
	"time"

	"github.com/lixianmin/logo"
)

/********************************************************************
created:    2024-10-23
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// ScanDefender 用于检测和防御基于IP的扫描行为
type ScanDefender struct {
	connectItems  map[string]*connectItem
	lastCleanupTs int64
}

type connectItem struct {
	lastConnectTs     int64
	connectTimestamps []int64
	isScanner         bool
}

// NewScanDefender 创建一个新的 ScanDefender 实例
func NewScanDefender() *ScanDefender {
	var sd = &ScanDefender{
		connectItems: make(map[string]*connectItem),
	}

	return sd
}

// IsScanner 记录一个新的连接并返回是否是扫描器
func (my *ScanDefender) IsScanner(ip string) bool {
	// 使用处理过的 ip，不包含端口号
	if ip == "" {
		return false
	}

	var item, exists = my.connectItems[ip]
	if exists && item.isScanner {
		return true
	}

	var now = time.Now().Unix()
	if !exists {
		item = &connectItem{}
		my.connectItems[ip] = item
	}

	item.lastConnectTs = now
	item.connectTimestamps = append(item.connectTimestamps, now)

	// 只保留最近10分钟的连接记录
	for len(item.connectTimestamps) > 0 && now-item.connectTimestamps[0] > 600 {
		item.connectTimestamps = item.connectTimestamps[1:]
	}

	// 检查是否是扫描器: 10分钟, 20次连接, 判定为扫描器
	var recentConnections = len(item.connectTimestamps)
	var isScanner = recentConnections >= 20
	if isScanner {
		item.isScanner = true
		item.connectTimestamps = nil
		logo.Info("[IsScanner()] find scanner, ip=%s, recentConnections=%d, len(connectItems)=%d", ip, recentConnections, len(my.connectItems))
	}

	// 检查是否需要清理过期的连接记录
	my.checkCleanup(now)
	return isScanner
}

// checkCleanup 清理过期的连接记录
func (my *ScanDefender) checkCleanup(now int64) {
	// 每1分钟清理一次过期的连接记录
	if now-my.lastCleanupTs > 60 {
		my.lastCleanupTs = now
		var before = len(my.connectItems)

		for ip, item := range my.connectItems {
			// 1小时内未连接的IP将被删除
			if now-item.lastConnectTs > 3600 {
				delete(my.connectItems, ip)
			}
		}

		var after = len(my.connectItems)
		logo.Info("[checkCleanup()] before=%d, after=%d", before, after)
	}
}
