package serde

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Packet struct {
	Kind      int32  // 自定义的类型从UserBase开始. 服务器push的消息是不需要handler的, 只需要route就可以 (当然, 为了优化也可以通过加一个handler引入一个kind)
	Route     []byte // kind >= RouteBase <=> Route有值; route需要是一个[]byte而不能是string, 因为需要在外围计算route的长度, 并赋值到kind, 计算前我们就已经拿到[]byte了
	RequestId int32  // 请求的rid, 用于client请求时定位response的handler
	Code      []byte // error code
	Data      []byte // 如果有error code, 则Data是error message; 否则Data是数据payload
}

// JsonHandshake handshake必须使用json做序列化与反序列化
type JsonHandshake struct {
	Nonce     int32    `json:"nonce"`
	Heartbeat float32  `json:"heartbeat"` // 心跳间隔. 单位: 秒
	Routes    []string `json:"routes"`    // 有序的routes, 其kinds值从Userdata(1000)有序增加; 只所以这么做并不是为了省流量, 而是unity3d的JsonUtility不支持反序列化Dictionary
	Serdes    []string `json:"serdes"`    // 服务器支持的序列化方法
}

type JsonHandshakeRe struct {
	Serde string `json:"serde"`
}
