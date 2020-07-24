package gonsole

import "github.com/lixianmin/gonsole/beans"

/********************************************************************
created:    2020-07-22
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type HtmlResponse struct {
	beans.BasicResponse
	Html string `json:"html"`
}
