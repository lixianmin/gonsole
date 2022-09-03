package codec

import "errors"

/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var (
	// ErrWrongPomeloPacketKind represents a wrong packet type.
	ErrWrongPomeloPacketKind = errors.New("wrong packet type")

	// ErrInvalidPomeloHeader represents an invalid header
	ErrInvalidPomeloHeader = errors.New("invalid header")

	// ErrPacketSizeExceed is the error used for encode/decode.
	ErrPacketSizeExceed = errors.New("codec: packet size exceed")
)
