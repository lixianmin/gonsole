/********************************************************************
 created:    2022-01-17
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

// 其实就是typescript中的enum编译之后的样子, 所以没有必要用enum
export const PacketType = {
    Handshake: 1,
    HandshakeAck: 2,
    Heartbeat: 3,
    Data: 4,
    Kick: 5,
}