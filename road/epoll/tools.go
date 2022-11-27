package epoll

/********************************************************************
created:    2020-09-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

//func checkReceivedMsgBytes(msgBytes []byte) error {
//	if len(msgBytes) < codec.HeaderSize {
//		return codec.ErrInvalidPomeloHeader
//	}
//
//	header := msgBytes[:codec.HeaderSize]
//	msgSize, _, err := codec.ParseHeader(header)
//	if err != nil {
//		return err
//	}
//
//	dataLen := len(msgBytes[codec.HeaderSize:])
//	if dataLen < msgSize {
//		return ifs.ErrReceivedMsgSmallerThanExpected
//	} else if dataLen > msgSize {
//		return ifs.ErrReceivedMsgBiggerThanExpected
//	}
//
//	return nil
//}
