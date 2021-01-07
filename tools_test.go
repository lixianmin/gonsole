package gonsole

import (
	"fmt"
	"testing"
)

/********************************************************************
created:    2021-01-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func TestNilToHtmlTable(t *testing.T) {
	var data interface{} = nil
	var result = ToHtmlTable(data)
	fmt.Println(result)
}
