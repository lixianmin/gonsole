'use strict'

/********************************************************************
 created:    2023-06-09
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export const PacketKind = {
    Handshake: 1,       // 连接建立后, 服务器主动发送handshake
    Heartbeat: 2,       // client定期发送心跳
    Kick: 3,            // server踢人
    UserBase: 10,       // 用户自定义的类型, 从这里开始
    RouteBase: 5000     // 当kind值 >= RouteBase时, 就意味着存储的是route字符串, route字符器长度=(kind - RouteBase)
}

