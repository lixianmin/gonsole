package intern

import (
	"sync"
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
	connectItemsLock sync.Mutex
	connectItems     map[string]*connectItem
	lastCleanupTs    int64
}

type connectItem struct {
	lastConnectTs int64
	isScanner     bool
	flashCloseNum int16 // 闪断计数器

	// connectTimestamps []int64
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

	var now = time.Now().UnixMilli()
	var item = my.fetchItem(ip)

	// 10分钟内没有连接的IP，重置闪断计数器
	// 这一条主要是针对那些偶尔连接一次的IP，防止误判. 但对于已经判定为扫描器的IP，仍然会保留isScanner=true
	if now-item.lastConnectTs > 10*60*1000 {
		item.flashCloseNum = 0
	}

	item.lastConnectTs = now
	var isScanner = item.isScanner

	if isScanner {
		return true
	}

	// item.connectTimestamps = append(item.connectTimestamps, now)

	// // 只保留最近10分钟的连接记录
	// for len(item.connectTimestamps) > 0 && now-item.connectTimestamps[0] > 600000 {
	// 	item.connectTimestamps = item.connectTimestamps[1:]
	// }

	// // 检查是否是扫描器: 10分钟, 20次连接, 判定为扫描器
	// var recentConnections = len(item.connectTimestamps)
	// isScanner = recentConnections >= 20
	// if isScanner {
	// 	item.isScanner = true
	// 	item.connectTimestamps = nil
	// 	logo.Info("[IsScanner()] find scanner, ip=%s, recentConnections=%d", ip, recentConnections)
	// }

	// 检查是否需要清理过期的连接记录
	my.checkCleanup(now)
	return isScanner
}

func (my *ScanDefender) OnConnectionClosed(ip string) {
	var item = my.getItem(ip)
	if item != nil {
		var now = time.Now().UnixMilli()
		// 1秒内关闭的连接，认为是闪断
		if now-item.lastConnectTs < 1000 {
			item.flashCloseNum++

			if item.flashCloseNum >= 20 {
				item.isScanner = true
				logo.Info("[OnConnectionClosed()] find scanner, ip=%s", ip)
			}
		}
	}
}

func (my *ScanDefender) fetchItem(ip string) *connectItem {
	my.connectItemsLock.Lock()

	var item, exists = my.connectItems[ip]
	if !exists {
		item = &connectItem{}
		my.connectItems[ip] = item
	}

	my.connectItemsLock.Unlock()
	return item
}

func (my *ScanDefender) getItem(ip string) *connectItem {
	if ip == "" {
		return nil
	}

	my.connectItemsLock.Lock()
	var item = my.connectItems[ip]
	my.connectItemsLock.Unlock()
	return item
}

// checkCleanup 清理过期的连接记录
func (my *ScanDefender) checkCleanup(now int64) {
	// 每1分钟清理一次过期的连接记录
	if now-my.lastCleanupTs > 60000 {
		my.connectItemsLock.Lock()
		defer my.connectItemsLock.Unlock()

		my.lastCleanupTs = now
		var before = len(my.connectItems)

		for ip, item := range my.connectItems {
			// 1小时内未连接的IP将被删除
			if now-item.lastConnectTs > 3600000 {
				delete(my.connectItems, ip)
			}
		}

		var after = len(my.connectItems)
		logo.Info("[checkCleanup()] before=%d, after=%d", before, after)
	}
}
