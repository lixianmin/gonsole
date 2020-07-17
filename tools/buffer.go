package tools

/********************************************************************
created:    2020-07-17
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func Itoa(buf *[]byte, i int, width int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || width > 1 {
		width--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

