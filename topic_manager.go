package gonsole

import (
	"github.com/lixianmin/gonsole/ifs"
	"sync"
)

/********************************************************************
created:    2024-02-03
author:     refactoring

TopicManager 负责 Topic 的注册和管理
Copyright (C) - All Rights Reserved
*********************************************************************/

type TopicManager struct {
	topics sync.Map
}

// NewTopicManager 创建新的主题管理器
func NewTopicManager() *TopicManager {
	return &TopicManager{}
}

// Register 注册主题
func (m *TopicManager) Register(topic *Topic) {
	if topic != nil && topic.Name != "" && topic.Interval > 0 && topic.BuildResponse != nil {
		m.topics.Store(topic.Name, topic)
		topic.start()
	}
}

// Get 获取指定名称的主题
func (m *TopicManager) Get(name string) *Topic {
	if box, ok := m.topics.Load(name); ok {
		if topic, ok := box.(*Topic); ok {
			return topic
		}
	}
	return nil
}

// GetAll 获取所有主题列表
func (m *TopicManager) GetAll() []ifs.Command {
	var list []ifs.Command
	m.topics.Range(func(key, value interface{}) bool {
		if topic, ok := value.(*Topic); ok {
			list = append(list, topic)
		}
		return true
	})
	return list
}

// RegisterBuiltins 注册内置主题
func (m *TopicManager) RegisterBuiltins(console *Console) {
	// 内置主题注册逻辑
	// 这里将在后续调用 console.registerBuiltinTopics 的逻辑
}
