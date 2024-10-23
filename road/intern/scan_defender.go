package intern

import (
	"sync"
	"sync/atomic"
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

	var item = my.fetchItem(ip)
	var isScanner = atomic.LoadInt32(&item.isScanner) > 0

	// 无论是否是scanner，都要更新连接时间
	var lastConnectTs = atomic.LoadInt64(&item.lastConnectTs)
	var now = time.Now().UnixMilli()
	atomic.StoreInt64(&item.lastConnectTs, now)
	if isScanner {
		return true
	}

	// 10分钟内没有新连接的，重置计数器
	if now-lastConnectTs > 10*60*1000 {
		item.ResetCounter()
	}

	// 当无效连接数超过10，判定为扫描器
	atomic.AddInt32(&item.connectingCounter, 1)
	if atomic.LoadInt32(&item.connectingCounter) >= 10 {
		atomic.StoreInt32(&item.isScanner, 1)
		logo.Info("[IsScanner()] find scanner by connectingCounter, ip=%s", ip)
		return true
	}

	// 检查是否需要清理过期的连接记录
	my.checkCleanup(now)
	return false
}

func (my *ScanDefender) OnSessionHandShaken(ip string) {
	var item = my.getItem(ip)
	if item != nil {
		// 当一家公司的出口ip都是同一个时, 重置计数器以防止误伤
		item.ResetCounter()
	}
}

func (my *ScanDefender) OnSessionClosed(ip string) {
	var item = my.getItem(ip)
	if item != nil {
		atomic.AddInt32(&item.connectingCounter, -1)

		var now = time.Now().UnixMilli()
		var lastConnectTs = atomic.LoadInt64(&item.lastConnectTs)

		if now-lastConnectTs < 1*1000 {
			var flashCloseNum = atomic.AddInt32(&item.flashCloseCounter, 1)
			if flashCloseNum >= 20 {
				atomic.StoreInt32(&item.isScanner, 1)
				logo.Info("[OnSessionClosed()] find scanner by flashCloseCounter, ip=%s", ip)
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
	if now-my.lastCleanupTs > 60*1000 {
		my.lastCleanupTs = now

		my.connectItemsLock.Lock()
		defer my.connectItemsLock.Unlock()

		var before = len(my.connectItems)
		for ip, item := range my.connectItems {
			var lastConnectTs = atomic.LoadInt64(&item.lastConnectTs)
			if now-lastConnectTs > 3600*1000 {
				delete(my.connectItems, ip)
			}
		}

		var after = len(my.connectItems)
		logo.Info("[checkCleanup()] before=%d, after=%d", before, after)
	}
}
