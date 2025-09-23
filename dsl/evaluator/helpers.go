package evaluator

import (
	"fmt"
	"onql/database/get"
	"onql/storemanager"
	"strings"
)

func GetTableData(db string, table string) ([]map[string]interface{}, error) {
	pks, err := get.GetAllPks(db, table)
	if err != nil {
		return nil, err
	}
	// fmt.Println(db,table)
	data, err := get.GetWithPKs(db, table, pks)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// func GetTableWithDataWithFilters(db string, table string, filters []string) ([]map[string]interface{}, error) {
// 	// loop filters and get by col value and at last union of pks then get with pks
// 	pks := make([]string, 0)
// 	for i, filter := range filters {
// 		if filter == "and" || filter == "or" {
// 			continue
// 		}
// 		cols := strings.Split(filter, ":")
// 		pk, err := get.GetPksFromIndex(db, table, cols[0]+":"+cols[1])
// 		if err != nil {
// 			return nil, err
// 		}
// 		// check is i and or or
// 		if i == 0 {
// 			pks = pk
// 			continue
// 		}
// 		// check if i-1 is and then union of pks then append otherwise direct appent
// 		if filters[i-1] == "and" {
// 			// union of pks and pk
// 			pks = get.Union(pks, pk)
// 		} else {
// 			pks = append(pks, pk...)
// 		}
// 	}

// 	data, err := get.GetWithPKs(db, table, pks)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return data, nil
// }

func GetTableWithDataWithFilters(db string, table string, filters []string) ([]map[string]interface{}, error) {
	if len(filters) == 0 {
		return GetTableData(db, table)
	}

	var (
		pksAccum   []string
		prevOp     string // "", "and", "or"
		expectExpr = true
	)

	for i, tok := range filters {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			return nil, fmt.Errorf("empty token at index %d", i)
		}

		if expectExpr {
			parts := strings.SplitN(tok, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("bad filter token %q at index %d; expected 'col:value'", tok, i)
			}
			col, val := parts[0], parts[1]

			pk, err := get.GetPksFromIndex(db, table, col+":"+val)
			if err != nil {
				return nil, err
			}

			if len(pksAccum) == 0 && prevOp == "" {
				// first expr
				pksAccum = dedupe(pk)
			} else {
				switch prevOp {
				case "and":
					pksAccum = intersect(pksAccum, pk)
				case "or":
					pksAccum = union(pksAccum, pk)
				default:
					// safety: if someone passes consecutive exprs w/o op, treat as AND
					pksAccum = intersect(pksAccum, pk)
				}
			}
			expectExpr = false
		} else {
			op := strings.ToLower(tok)
			if op != "and" && op != "or" {
				return nil, fmt.Errorf("expected logical operator 'and' or 'or' at index %d, got %q", i, tok)
			}
			prevOp = op
			expectExpr = true
		}
	}

	// If the last token was an operator (dangling), that’s an input error.
	if expectExpr {
		return nil, fmt.Errorf("filters end with operator %q; missing right-hand expression", prevOp)
	}

	pksAccum = dedupe(pksAccum)
	if len(pksAccum) == 0 {
		return []map[string]interface{}{}, nil
	}
	return get.GetWithPKs(db, table, pksAccum)
}

func GetRelatedTableData(db string, relation storemanager.Relation, value string) ([]map[string]interface{}, error) {
	//two probelems pending first original col name table name and db name not alias second mtm through table thirds in oto and mto case send dict not array
	if relation.Type == "mtm" {
		return GetMTMRelatedTabledData(db, relation, value)
	}
	cols := strings.Split(relation.FKField, ":")
	pks, err := get.GetPksFromIndex(db, relation.Entity, cols[1]+":"+value)
	if err != nil {
		return nil, err
	}
	data, err := get.GetWithPKs(db, relation.Entity, pks)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetMTMRelatedTabledData(db string, relation storemanager.Relation, value string) ([]map[string]interface{}, error) {
	cols := strings.Split(relation.FKField, ":")
	pks, err := get.GetPksFromIndex(db, relation.Through, cols[1]+":"+value)
	if err != nil {
		return nil, err
	}
	data, err := get.GetWithPKs(db, relation.Through, pks)
	if err != nil {
		return nil, err
	}
	values := make([]string, 0)
	for _, item := range data {
		if val, ok := item[cols[2]]; ok {
			values = append(values, val.(string))
		}
	}
	return GetTableDataWithColValues(db, relation.Entity, cols[3], values)
}

func GetTableDataWithColValues(db string, table string, col string, values []string) ([]map[string]interface{}, error) {
	pksOuter := make([]string, 0)
	for _, value := range values {
		pks, err := get.GetPksFromIndex(db, table, col+":"+value)
		if err != nil {
			return nil, err
		}
		pksOuter = append(pksOuter, pks...)
	}
	data, err := get.GetWithPKs(db, table, pksOuter)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// func (e *Evaluator) GetDataFromVar(varName string) (interface{}, error) {
// 	value, ok := e.Memory[varName]
// 	if !ok {
// 		return nil, fmt.Errorf("variable not found: %s", varName)
// 	}
// 	return value, nil
// }

// ---------------------- set helpers ----------------------

func union(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	for _, x := range a {
		seen[x] = struct{}{}
	}
	for _, x := range b {
		seen[x] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}

func intersect(a, b []string) []string {
	if len(a) == 0 || len(b) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(a))
	for _, x := range a {
		seen[x] = struct{}{}
	}
	out := make([]string, 0)
	for _, x := range b {
		if _, ok := seen[x]; ok {
			out = append(out, x)
		}
	}
	// optional dedupe if b had dups (rare)
	if len(out) > 1 {
		tmp := make(map[string]struct{}, len(out))
		uniq := out[:0]
		for _, x := range out {
			if _, ok := tmp[x]; !ok {
				tmp[x] = struct{}{}
				uniq = append(uniq, x)
			}
		}
		out = uniq
	}
	return out
}

// dedupe keeps order roughly arbitrary; if you need stable sort, sort.Strings after
func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, x := range in {
		if _, ok := seen[x]; !ok {
			seen[x] = struct{}{}
			out = append(out, x)
		}
	}
	return out
}
