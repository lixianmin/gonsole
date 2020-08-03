package beans

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/lixianmin/gonsole/tools"
	"io"
	"regexp"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
)

/********************************************************************
created:    2020-08-02
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type DeadItem struct {
	Count    int
	Text     string
	waitTime int
}

func DeadlockDetect() string {
	var title = ""
	var list = make([]string, 0, 8)
	var itemMap = make(map[string]*DeadItem, 16)

	// 匹配title
	var titlePattern, _ = regexp.Compile(`goroutine.*\[.*?(\d+) minutes\]:`)

	// 匹配一个调用方法
	//var funcPattern, _ = regexp.Compile(`\s*(.*)\(.*\)`)

	var err = readPProfGoroutineByLine(func(line string) {
		if strings.HasPrefix(line, "goroutine") {
			// 此分支是一条记录的开始
			title = line
		} else if strings.TrimSpace(line) == "" {
			// 此分支是一条记录的结束
			body := strings.Join(list, "<br>")
			body = strings.ReplaceAll(body, "\t", "&nbsp;&nbsp;")

			list = list[:0]
			item, ok := itemMap[body]
			if !ok {
				item = &DeadItem{}
				itemMap[body] = item
			}

			item.Count += 1
			// title = "goroutine 105 [IO wait, 17 minutes]:"
			match := titlePattern.FindStringSubmatch(title)
			if match != nil {
				item.Text = title + "<br>" + body
				waitTime, _ := strconv.Atoi(match[1])
				if waitTime > item.waitTime {
					item.waitTime = waitTime
				}
			} else if item.Text == "" {
				item.Text = title + "<br>" + body
			}
		} else {
			// 此分支处理调用栈的数据行
			//match := funcPattern.FindStringSubmatch(line)
			//if match != nil {
			//	list = append(list, match[1])
			//} else {
			//	list = append(list, line)
			//}
			list = append(list, line)
		}
	})

	if err != nil {
		return err.Error()
	}

	items := make([]*DeadItem, 0, len(itemMap))
	for _, v := range itemMap {
		if v.waitTime > 0 {
			items = append(items, v)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.waitTime < b.waitTime {
			return false
		} else if a.waitTime > b.waitTime {
			return true
		}

		return a.Count > b.Count
	})

	return tools.ToHtmlTable(items)
}

func readPProfGoroutineByLine(handler func(line string)) error {
	const name = "goroutine"
	const debug = 2
	p := pprof.Lookup(name)
	if p == nil {
		return fmt.Errorf("can not find pprof type: %q", name)
	}

	var buff bytes.Buffer
	buff.Grow(4096)
	err := p.WriteTo(&buff, debug)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(&buff)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if io.EOF == err {
				break
			}

			return err
		}

		handler(line)
	}

	return nil
}
