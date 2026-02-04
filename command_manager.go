package gonsole

import (
	"github.com/lixianmin/gonsole/ifs"
	"sync"
)

/********************************************************************
created:    2024-02-03
author:     refactoring

CommandManager 负责命令的注册和管理
Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandManager struct {
	commands sync.Map
}

// NewCommandManager 创建新的命令管理器
func NewCommandManager() *CommandManager {
	return &CommandManager{}
}

// Register 注册命令
func (m *CommandManager) Register(cmd *Command) {
	if cmd != nil && cmd.Name != "" {
		m.commands.Store(cmd.Name, cmd)
	}
}

// Get 获取指定名称的命令
func (m *CommandManager) Get(name string) ifs.Command {
	if box, ok := m.commands.Load(name); ok {
		if cmd, ok := box.(*Command); ok {
			return cmd
		}
	}
	return nil
}

// GetAll 获取所有命令列表
func (m *CommandManager) GetAll() []ifs.Command {
	var list []ifs.Command
	m.commands.Range(func(key, value any) bool {
		if cmd, ok := value.(*Command); ok {
			list = append(list, cmd)
		}
		return true
	})

	// 添加内置的 request 命令
	list = append(list, &Command{
		Name:    "request",
		Example: `request console.command {"command":"help"}`,
		Note:    "模拟直接发送请求",
		Flag:    flagBuiltin,
	})

	return list
}

// RegisterBuiltins 注册内置命令
func (m *CommandManager) RegisterBuiltins(port int, console *Console) {
	// 内置命令注册逻辑
	// 这里将在后续调用 console.registerBuiltinCommands 的逻辑
}
