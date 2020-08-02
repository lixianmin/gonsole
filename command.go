package gonsole

/********************************************************************
created:    2020-06-04
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Command struct {
	Name     string                               // 名称
	Note     string                               // 描述
	IsPublic bool                                 // 非public方法需要登陆
	Handler  func(client *Client, args [] string) // 处理方法

	isBuiltin bool // 是否为内置命令，排序时内置命令排在前面
}

func (cmd *Command) GetName() string {
	return cmd.Name
}

func (cmd *Command) GetNote() string {
	return cmd.Note
}

func (cmd *Command) CheckPublic() bool {
	return cmd.IsPublic
}

func (cmd *Command) CheckBuiltin() bool {
	return cmd.isBuiltin
}
