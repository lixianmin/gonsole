package tools

import (
	"fmt"
	"testing"
	"time"
)

/********************************************************************
created:    2020-07-24
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type TestHtml struct {
	Height     float64
	Age        int
	Name       string
	CreateTime time.Time
	IsMan      bool
	Nil        *int
	SliceNil   []int
	Err        error
}

func TestToHtmlTableStruct(t *testing.T) {
	var item = TestHtml{
		Height:     10.29,
		Age:        98,
		Name:       "pet",
		CreateTime: time.Now(),
		IsMan:      true,
		Nil:        nil,
		SliceNil:   nil,
		Err:        fmt.Errorf("i_am_error"),
	}

	var html = ToHtmlTable(item)
	fmt.Println(html)
}

func TestToHtmlTableSlice(t *testing.T) {
	var list = []TestHtml{{
		Height:     10.29,
		Age:        98,
		Name:       "pet",
		CreateTime: time.Now(),
		IsMan:      false,
	}, {
		Height:     5.6,
		Age:        85,
		Name:       "滴滴",
		CreateTime: time.Now(),
		IsMan:      false,
	}}

	var html = ToHtmlTable(list)
	fmt.Println(html)
}

func TestToHtmlTableStringSlice(t *testing.T) {
	var list = []interface{}{123.0, true, 1, "world"}

	var html = ToHtmlTable(list)
	fmt.Println(html)
}

func TestToHtmlTablePointer(t *testing.T) {
	var list = []*TestHtml{{
		Height:     10.29,
		Age:        98,
		Name:       "pet",
		CreateTime: time.Now(),
	}, {
		Height:     5.6,
		Age:        85,
		Name:       "滴滴",
		CreateTime: time.Now(),
	}}

	var html = ToHtmlTable(list)
	fmt.Println(html)
}
