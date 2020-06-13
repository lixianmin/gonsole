package tools

import (
	"os"
)

/********************************************************************
created:    2020-06-02
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func IsPathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func EnsureDir(dirname string) error {
	if _, err := os.Stat(dirname); err != nil {
		err = os.MkdirAll(dirname, os.ModePerm)
		return err
	}

	return nil
}