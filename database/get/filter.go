package get

import (
	"errors"
	"onql/database"
	"onql/storemanager"
	"strings"
	"sync"
)

type FilterOp string

const (
	OpEq    FilterOp = "="
	OpNeq   FilterOp = "!="
	OpGt    FilterOp = ">"
	OpGte   FilterOp = ">="
	OpLt    FilterOp = "<"
	OpLte   FilterOp = "<="
	OpIn    FilterOp = "in"
	OpNotIn FilterOp = "not_in"
	OpAnd   FilterOp = "and"
	OpOr    FilterOp = "or"
)

type Filter struct {
	Column    string
	Operator  FilterOp
	Operand2  string
	SubFilter []Filter // for nested and/or groups
	Logic     FilterOp // "and" or "or"
}

func GetPksFromFilters(db, table string, filters []Filter) ([]string, error) {
	if !database.IsDatabaseExists(db) {
		return nil, errors.New("database does not exist")
	}
	if !database.IsTableExists(db, table) {
		return nil, errors.New("table does not exist")
	}

	schema := database.FullSchema[db][table]

	pkSet, err := evalFilterGroup(db, table, schema, filters, OpAnd)
	if err != nil {
		return nil, err
	}

	return mapKeys(pkSet), nil
}

func evalFilterGroup(db, table string, schema map[string]map[string]string, filters []Filter, logic FilterOp) (map[string]bool, error) {
	var wg sync.WaitGroup
	resCh := make(chan map[string]bool, len(filters))
	errCh := make(chan error, len(filters))

	for _, f := range filters {
		f := f
		wg.Add(1)
		go func() {
			defer wg.Done()
			var set map[string]bool
			var err error

			if len(f.SubFilter) > 0 {
				set, err = evalFilterGroup(db, table, schema, f.SubFilter, f.Logic)
			} else {
				colType := schema[f.Column]["type"]
				set, err = runSingleFilter(db, table, f, colType)
			}

			if err != nil {
				errCh <- err
				return
			}
			resCh <- set
		}()
	}

	wg.Wait()
	close(resCh)
	close(errCh)

	if len(errCh) > 0 {
		return nil, <-errCh
	}

	sets := make([]map[string]bool, 0, len(filters))
	for s := range resCh {
		sets = append(sets, s)
	}

	if len(sets) == 0 {
		return nil, nil
	}

	result := sets[0]
	for i := 1; i < len(sets); i++ {
		if logic == OpOr {
			result = union(result, sets[i])
		} else {
			result = intersect(result, sets[i])
		}
	}
	return result, nil
}

func runSingleFilter(db, table string, filter Filter, colType string) (map[string]bool, error) {
	switch filter.Operator {
	case OpIn, OpNotIn:
		values := strings.Split(filter.Operand2, ",")
		seen := make(map[string]bool)
		for _, val := range values {
			val = strings.TrimSpace(val)
			tmp := storemanager.FilterPksByIndex(db, table, filter.Column, "=", colType, val)
			for _, pk := range tmp {
				seen[pk] = true
			}
		}
		if filter.Operator == OpIn {
			return seen, nil
		}
		// NOT IN
		all := storemanager.FilterPksByIndex(db, table, filter.Column, ">=", colType, "")
		for _, pk := range all {
			if seen[pk] {
				delete(seen, pk)
			}
		}
		return seen, nil

	default:
		pks := storemanager.FilterPksByIndex(db, table, filter.Column, string(filter.Operator), colType, filter.Operand2)
		return toSet(pks), nil
	}
}
