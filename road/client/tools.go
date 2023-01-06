package client

import (
	"io"
	"net"
)

/********************************************************************
created:    2023-01-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func readConnection(buffer io.Writer, conn net.Conn) error {
	var data [512]byte // 这种方式声明的data是一个实际存储在栈上的array
	for {
		var n, err = conn.Read(data[:])
		if err != nil {
			return err
		}

		if _, err2 := buffer.Write(data[:n]); err2 != nil {
			return err2
		}

		if n < len(data) {
			return nil
		}
	}
}
