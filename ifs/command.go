package ifs

/********************************************************************
created:    2020-07-24
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Command interface {
	GetName() string
	GetExample() string
	GetNote() string
	IsPublic() bool
	IsInvisible() bool
}
