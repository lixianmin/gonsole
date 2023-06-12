package serde

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Packet struct {
	Kind      int32  // 自定义的类型从UserDefined开始
	RequestId int32  // 请求的rid, 用于client请求时定位response的handler
	Code      []byte // error code
	Data      []byte // 如果有error code, 则Data是error message; 否则Data是数据payload
}

type HandshakeInfo struct {
	Nonce     int32    `json:"nonce"`
	Heartbeat float32  `json:"heartbeat"` // 心跳间隔. 单位: 秒
	Routes    []string `json:"routes"`    // 有序的routes, 其kinds值从Userdata(1000)有序增加; 只所以这么做并不是为了省流量, 而是unity3d的JsonUtility不支持反序列化Dictionary
}
