package ifs

/********************************************************************
created:    2020-09-01
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/
type Server interface {
	GetCommands() []Command
}
