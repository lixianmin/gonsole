package epoll

import (
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road/codec"
)

/********************************************************************
created:    2020-09-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func checkReceivedMsgBytes(msgBytes []byte) error {
	if len(msgBytes) < codec.HeadLength {
		return ifs.ErrInvalidPomeloHeader
	}

	header := msgBytes[:codec.HeadLength]
	msgSize, _, err := codec.ParseHeader(header)
	if err != nil {
		return err
	}

	dataLen := len(msgBytes[codec.HeadLength:])
	if dataLen < msgSize {
		return ifs.ErrReceivedMsgSmallerThanExpected
	} else if dataLen > msgSize {
		return ifs.ErrReceivedMsgBiggerThanExpected
	}

	return nil
}
