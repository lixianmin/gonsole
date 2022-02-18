/********************************************************************
 created:    2022-01-17
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export enum PacketType {
    Handshake = 1,
    HandshakeAck = 2,
    Heartbeat = 3,
    Data = 4,
    Kick = 5,
}