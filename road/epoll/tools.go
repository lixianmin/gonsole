package epoll

import (
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road/codec"
	"github.com/lixianmin/got/iox"
)

/********************************************************************
created:    2020-09-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func checkReceivedMsgBytes(msgBytes []byte) error {
	if len(msgBytes) < codec.HeaderSize {
		return codec.ErrInvalidPomeloHeader
	}

	header := msgBytes[:codec.HeaderSize]
	msgSize, _, err := codec.ParseHeader(header)
	if err != nil {
		return err
	}

	dataLen := len(msgBytes[codec.HeaderSize:])
	if dataLen < msgSize {
		return ifs.ErrReceivedMsgSmallerThanExpected
	} else if dataLen > msgSize {
		return ifs.ErrReceivedMsgBiggerThanExpected
	}

	return nil
}

func onReceiveMessage(input *iox.Buffer, onReadHandler OnReadHandler) error {
	var headSize = codec.HeaderSize
	var data = input.Bytes()

	// 像heartbeat之类的协议，有可能只有head没有body，所以需要使用>=
	for len(data) >= headSize {
		var header = data[:headSize]
		bodySize, _, err := codec.ParseHeader(header)
		if err != nil {
			onReadHandler(nil, err)
			return err
		}

		var totalSize = headSize + bodySize
		if len(data) < totalSize {
			return nil
		}

		// 这里每次新建的frameData目前是省不下的, 原因是writeMessage()方法会把这个slice写到chan中并由另一个goroutine使用
		//var frameData = make([]byte, totalSize)
		//copy(frameData, data[:totalSize])
		// onReadHandler()会把data[]中的数据copy走，因此不再需要新生成一个frameData
		onReadHandler(data[:totalSize], nil)

		input.Next(totalSize)
		data = input.Bytes()
	}

	input.Tidy()
	return nil
}
