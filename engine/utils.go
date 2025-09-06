package engine

import (
	"strconv"
	"strings"
)

func compare(a, b string, op string, valType string) bool {
	switch op {
	case "in", "not in":
		// `b` should be a comma-separated list for `in` / `not in`
		bList := strings.Split(b, ",")
		match := false
		for _, item := range bList {
			item = strings.TrimSpace(item)
			if valType == "number" || valType == "timestamp" {
				af, err1 := strconv.ParseFloat(a, 64)
				bf, err2 := strconv.ParseFloat(item, 64)
				if err1 == nil && err2 == nil && af == bf {
					match = true
					break
				}
			} else if valType == "string" {
				if a == item {
					match = true
					break
				}
			}
		}
		if op == "in" {
			return match
		}
		return !match
	}

	switch valType {
	case "number", "timestamp":
		af, err1 := strconv.ParseFloat(a, 64)
		bf, err2 := strconv.ParseFloat(b, 64)
		if err1 != nil || err2 != nil {
			return false
		}
		switch op {
		case "<":
			return af < bf
		case "<=":
			return af <= bf
		case ">":
			return af > bf
		case ">=":
			return af >= bf
		case "=":
			return af == bf
		case "!=":
			return af != bf
		}
	case "string":
		switch op {
		case "<":
			return a < b
		case "<=":
			return a <= b
		case ">":
			return a > b
		case ">=":
			return a >= b
		case "=":
			return a == b
		case "!=":
			return a != b
		}
	}
	return false
}
