package tools

import (
	"bytes"
	"encoding/json"
)

/********************************************************************
created:    2020-07-04
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func MarshalUnescape(v interface{}) ([]byte, error) {
	var buffer = bytes.NewBuffer([]byte{})
	var encoder = json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	var err = encoder.Encode(v)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
