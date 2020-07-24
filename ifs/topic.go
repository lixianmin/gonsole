package ifs

/********************************************************************
created:    2020-07-24
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Topic interface {
	GetName() string
	GetNote() string
	CheckPublic() bool
}
