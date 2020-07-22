package gonsole

/********************************************************************
created:    2020-07-22
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type HtmlResponse struct {
	BasicResponse
	Html string `json:"html"`
}
