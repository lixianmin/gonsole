package client

import "errors"

/********************************************************************
created:    2023-01-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

const (
	StateNone      = 0
	StateHandshake = 1
	StateConnected = 2
)

var ErrKicked = errors.New("got kick packet from the server! disconnecting~")
