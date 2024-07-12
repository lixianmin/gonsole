package serde

import "github.com/lixianmin/got/convert"

/********************************************************************
created:    2023-06-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type JsonSerde struct {
}

func (s *JsonSerde) Serialize(v any) ([]byte, error) {
	if bts, ok := v.([]byte); ok {
		return bts, nil
	}

	if v != nil {
		return convert.ToJsonE(v)
	}

	return nil, nil
}

func (s *JsonSerde) Deserialize(data []byte, v any) error {
	return convert.FromJsonE(data, v)
}
