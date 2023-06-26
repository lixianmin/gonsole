package serde

/********************************************************************
created:    2023-06-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Serde interface {
	Serialize(v interface{}) ([]byte, error)
	Deserialize(data []byte, v interface{}) error
	GetName() string
}
