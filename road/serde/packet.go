package serde

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Packet struct {
	Kind int32  // 自定义的类型从UserDefined开始
	Code []byte // error code
	Data []byte // 如果有error code, 则Data是error message; 否则Data是数据payload
}
