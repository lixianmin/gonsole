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

type DetectItem struct {
	Count    int
	Text     string
	waitTime int
	title    string
}

func DeadlockDetect(args []string, deadlockIgnores []string) string {
	deadlockIgnores = checkDeadlockDetectArgs(args, deadlockIgnores)

	var isIgnored = false
	var title = ""
	var list = make([]string, 0, 8)
	var itemMap = make(map[string]*DetectItem, 16)

	// 匹配title
	var titlePattern, _ = regexp.Compile(`goroutine.*\[.*?(\d+) minutes\]:`)

	// 匹配一个调用方法
	//var funcPattern, _ = regexp.Compile(`\s*(.*)\(.*\)`)

	var err = readPProfGoroutineByLine(func(line string) {
		if strings.HasPrefix(line, "goroutine") {
			// 此分支是一条记录的开始
			isIgnored = false
			title = line
			list = list[:0]
		} else if !isIgnored {
			if strings.TrimSpace(line) == "" {
				// 此分支是一条记录的结束
				body := strings.Join(list, "<br>")
				body = strings.ReplaceAll(body, "\t", "&nbsp;&nbsp;&nbsp;&nbsp;")

				item, ok := itemMap[body]
				if !ok {
					item = &DetectItem{}
					itemMap[body] = item
				}

				item.Count += 1
				// title = "goroutine 105 [IO wait, 17 minutes]:"
				match := titlePattern.FindStringSubmatch(title)
				if match != nil {
					item.Text = title + "<br>" + body
					waitTime, _ := strconv.Atoi(match[1])
					if waitTime >= item.waitTime {
						item.waitTime = waitTime
						item.title = title
					}
				} else if item.Text == "" {
					item.Text = title + "<br>" + body
				}
			} else {
				// 此分支处理调用栈的数据行
				if len(list) == 0 && isDeadlockIgnored(deadlockIgnores, line) {
					isIgnored = true
				}

				list = append(list, line)
			}
		}
	})

	if err != nil {
		return err.Error()
	}

	items := make([]*DetectItem, 0, len(itemMap))
	for _, v := range itemMap {
		if v.waitTime > 0 {
			items = append(items, v)
		}
	}

	sortDetectItems(items)
	return tools.ToHtmlTable(items)
}

func isDeadlockIgnored(deadlockIgnores []string, line string) bool {
	for _, item := range deadlockIgnores {
		if strings.HasPrefix(line, item) {
			return true
		}
	}

	return false
}

func checkDeadlockDetectArgs(args []string, deadlockIgnores []string) []string {
	// deadlockIgnores
	if len(deadlockIgnores) == 0 {
		deadlockIgnores = append(deadlockIgnores, "internal/poll.runtime_pollWait(")
	}
	sort.Strings(deadlockIgnores)

	// args
	if len(args) >= 2 && strings.Contains(args[1], "-a") {
		deadlockIgnores = make([]string, 0)
	}

	return deadlockIgnores
}

func sortDetectItems(items []*DetectItem) {
	sort.Slice(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.waitTime < b.waitTime {
			return true
		} else if a.waitTime > b.waitTime {
			return false
		}

		if a.Count < b.Count {
			return true
		} else if a.Count > b.Count {
			return false
		}

		return a.title < b.title
	})
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
