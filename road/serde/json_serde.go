package serde

import "github.com/lixianmin/got/convert"

/********************************************************************
created:    2023-06-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type JsonSerde struct {
}

func (s *JsonSerde) Serialize(v interface{}) ([]byte, error) {
	if v != nil {
		return convert.ToJsonE(v)
	}

	return nil, nil
}

func (s *JsonSerde) Deserialize(data []byte, v interface{}) error {
	return convert.FromJsonE(data, v)
}

func (s *JsonSerde) GetName() string {
	return "json"
}
