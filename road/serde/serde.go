package serde

/********************************************************************
created:    2023-06-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Serde interface {
	Serialize(v any) ([]byte, error)
	Deserialize(data []byte, v any) error
	GetName() string
}
