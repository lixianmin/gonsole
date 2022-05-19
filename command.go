package gonsole

import "github.com/lixianmin/logo"

/********************************************************************
created:    2020-06-04
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Command struct {
	logo.Flag                                                        // command的flag
	Name      string                                                 // 名称
	Note      string                                                 // 描述
	Handler   func(client *Client, args []string) (*Response, error) // 处理方法
}

func (cmd *Command) GetName() string {
	return cmd.Name
}

func (cmd *Command) GetNote() string {
	return cmd.Note
}

func (cmd *Command) IsBuiltin() bool {
	return cmd.HasFlag(FlagBuiltin)
}

func (cmd *Command) IsPublic() bool {
	return cmd.HasFlag(FlagPublic)
}

func (cmd *Command) IsInvisible() bool {
	return cmd.HasFlag(FlagInvisible)
}

func (cmd *Command) Run(client *Client, args []string) (*Response, error) {
	return cmd.Handler(client, args)
}
