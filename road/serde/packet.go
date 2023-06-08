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

type HandshakeInfo struct {
	Nonce      int32            `json:"nonce"`
	Heartbeat  float32          `json:"heartbeat"` // 心跳间隔. 单位: 秒
	RouteKinds map[string]int32 `json:"route_kinds"`
}
