package get

import "onql/utils"

func toSet(list []string) map[string]bool {
	set := make(map[string]bool)
	for _, v := range list {
		set[v] = true
	}
	return set
}

func mapKeys(m map[string]bool) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func union(a, b map[string]bool) map[string]bool {
	res := make(map[string]bool)
	for k := range a {
		res[k] = true
	}
	for k := range b {
		res[k] = true
	}
	return res
}

func intersect(a, b map[string]bool) map[string]bool {
	res := make(map[string]bool)
	for k := range a {
		if b[k] {
			res[k] = true
		}
	}
	return res
}

func buildIndexKey(db, table, column, value string) string {
	return "index:" + db + ":" + table + ":" + column + ":" + utils.ToString(value)
}
