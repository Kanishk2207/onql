package evaluator

import (
	"fmt"
	"onql/dsl/parser"
	"sort"
	"strconv"
	"strings"
	"time"
)

// var AggrRegistry = map[string]map[string]fin{
// 	"_sum":      ,
// 	"_count":    {"LIST": "NUMBER"},
// 	"_avg":      {"LIST": "NUMBER"},
// 	"_min":      {"LIST": "NUMBER"},
// 	"_max":      {"LIST": "NUMBER"},
// 	"_distinct": {"LIST": "NUMBER"},
// }

var AggrRegistry = map[string]func(stmt *parser.Statement, data interface{}, aggrObj parser.Aggr, e *Evaluator) error{
	// "_sum":  _sum,
	// "_asc":  _asc,
	// "_desc": _desc,

	"_sum":    _sum,
	"_count":  _count,
	"_avg":    _avg,
	"_min":    _min,
	"_max":    _max,
	"_unique": _unique,
	"_date":   _date,
	"_asc":    _asc,
	"_desc":   _desc,
}

func (e *Evaluator) EvalAggr() error {
	stmt := e.Plan.NextStatement(true)
	if stmt.Operation != parser.OpAggregateReduce {
		return fmt.Errorf("expect aggregate reduce but got %s", stmt.Operation)
	}
	// prevStmt := e.Plan.PrevStatement(false)
	data := e.Memory[stmt.Sources[0].SourceValue]

	aggrObj := stmt.Expressions.(parser.Aggr)
	err := AggrRegistry[aggrObj.Name](stmt, data, aggrObj, e)
	if err != nil {
		return err
	}
	// switch aggrObj.Name {
	// case "_sum":
	// 	return _sum(stmt, data, aggrObj, e)
	// }
	return nil
}

func _sum(stmt *parser.Statement, data interface{}, aggrObj parser.Aggr, e *Evaluator) error {
	total := 0.0
	list := make([]float64, 0)
	if stmt.Meta["input_type"] == "TABLE" {
		for _, row := range data.([]map[string]interface{}) {
			list = append(list, row[aggrObj.Args[0]].(float64))
		}
	} else {
		list = data.([]float64)
	}
	for _, v := range list {
		total += v
	}
	// e.Memory[stmt.Name] = total
	// e.Memory[stmt.Name+"_meta_type"] = "NUMBER"
	e.SetMemoryValue(stmt.Name, total)
	return nil
}

// _asc orders data ASC. For TABLE, it sorts by args left→right: if the first key ties,
// it falls through to the next, like SQL ORDER BY col1, col2, ...
func _asc(stmt *parser.Statement, data interface{}, aggrObj parser.Aggr, e *Evaluator) error {
	switch v := data.(type) {

	// ---------- TABLE ----------
	case []map[string]interface{}:
		if len(aggrObj.Args) == 0 {
			return fmt.Errorf("sort: missing sort key(s)")
		}
		keys := make([]string, len(aggrObj.Args))
		for i, a := range aggrObj.Args {
			// s, ok := a.(string)
			if a == "" {
				return fmt.Errorf("sort: key %d must be non-empty string, got %T", i, a)
			}
			keys[i] = a
		}

		sort.SliceStable(v, func(i, j int) bool {
			ri, rj := v[i], v[j]
			for _, k := range keys {
				vi, vj := ri[k], rj[k]

				// nils last
				if vi == nil && vj == nil {
					continue
				}
				if vi == nil {
					return false
				}
				if vj == nil {
					return true
				}

				// numeric compare if both numeric, else string compare
				if fi, ok := asFloat64(vi); ok {
					if fj, ok := asFloat64(vj); ok {
						if fi < fj {
							return true
						}
						if fi > fj {
							return false
						}
						continue
					}
				}
				si := fmt.Sprint(vi)
				sj := fmt.Sprint(vj)
				if si < sj {
					return true
				}
				if si > sj {
					return false
				}
				// equal on this key → check next key
			}
			return false // completely equal
		})

		e.SetMemoryValue(stmt.Name, v)
		return nil

	// ---------- LIST (numbers) ----------
	case []float64:
		sort.Float64s(v)
		e.SetMemoryValue(stmt.Name, v)
		return nil

	// ---------- LIST (strings) ----------
	case []string:
		sort.Strings(v)
		e.SetMemoryValue(stmt.Name, v)
		return nil

	default:
		return fmt.Errorf("sort: expected TABLE ([]map[string]interface{}) or LIST ([]string/[]float64), got %T", data)
	}
}

// _sortDesc orders data in DESC order.
// TABLE: sorts by args left→right (col1 DESC, then col2 DESC, ...).
// LIST: sorts []float64 or []string in descending order.
// []interface{} is not supported for lists.
func _desc(stmt *parser.Statement, data interface{}, aggrObj parser.Aggr, e *Evaluator) error {
	switch v := data.(type) {

	// ---------- TABLE ----------
	case []map[string]interface{}:
		if len(aggrObj.Args) == 0 {
			return fmt.Errorf("sortDesc: missing sort key(s)")
		}
		keys := make([]string, len(aggrObj.Args))
		for i, a := range aggrObj.Args {
			// s, ok := a.(string)
			if a == "" {
				return fmt.Errorf("sortDesc: key %d must be non-empty string, got %T", i, a)
			}
			keys[i] = a
		}

		sort.SliceStable(v, func(i, j int) bool {
			ri, rj := v[i], v[j]
			for _, k := range keys {
				vi, vj := ri[k], rj[k]

				// nils last (same as ASC)
				if vi == nil && vj == nil {
					continue
				}
				if vi == nil {
					return false
				}
				if vj == nil {
					return true
				}

				// numeric compare if both numeric; else string compare
				if fi, ok := asFloat64(vi); ok {
					if fj, ok := asFloat64(vj); ok {
						if fi > fj { // DESC
							return true
						}
						if fi < fj {
							return false
						}
						continue
					}
				}
				si := fmt.Sprint(vi)
				sj := fmt.Sprint(vj)
				if si > sj { // DESC
					return true
				}
				if si < sj {
					return false
				}
				// equal on this key → check next key
			}
			return false // completely equal
		})

		e.SetMemoryValue(stmt.Name, v)
		return nil

	// ---------- LIST (numbers) ----------
	case []float64:
		sort.Sort(sort.Reverse(sort.Float64Slice(v)))
		e.SetMemoryValue(stmt.Name, v)
		return nil

	// ---------- LIST (strings) ----------
	case []string:
		sort.Sort(sort.Reverse(sort.StringSlice(v)))
		e.SetMemoryValue(stmt.Name, v)
		return nil

	default:
		return fmt.Errorf("sortDesc: expected TABLE ([]map[string]interface{}) or LIST ([]string/[]float64), got %T", data)
	}
}

func _count(stmt *parser.Statement, data interface{}, aggrObj parser.Aggr, e *Evaluator) error {
	var n int
	switch v := data.(type) {
	case []float64:
		n = len(v)
	case []string:
		n = len(v)
	case []bool:
		n = len(v)
	case []map[string]interface{}:
		// supports count on table rows as well
		n = len(v)
	default:
		return fmt.Errorf("_count: unsupported input %T", data)
	}
	e.SetMemoryValue(stmt.Name, float64(n))
	return nil
}

func _avg(stmt *parser.Statement, data interface{}, aggrObj parser.Aggr, e *Evaluator) error {
	sum := 0.0
	cnt := 0.0
	switch t := data.(type) {
	case []float64:
		for _, v := range t {
			sum += v
			cnt++
		}
	case []map[string]interface{}:
		if len(aggrObj.Args) == 0 {
			return fmt.Errorf("_avg: missing column name")
		}
		col := aggrObj.Args[0]
		for _, r := range t {
			if f, ok := asFloat64(r[col]); ok {
				sum += f
				cnt++
			}
		}
	default:
		return fmt.Errorf("_avg: unsupported input %T", data)
	}
	if cnt == 0 {
		e.SetMemoryValue(stmt.Name, 0.0)
		return nil
	}
	e.SetMemoryValue(stmt.Name, sum/cnt)
	return nil
}

func _min(stmt *parser.Statement, data interface{}, aggrObj parser.Aggr, e *Evaluator) error {
	minSet := false
	minVal := 0.0

	switch t := data.(type) {
	case []float64:
		for _, v := range t {
			if !minSet || v < minVal {
				minVal = v
				minSet = true
			}
		}
	case []map[string]interface{}:
		if len(aggrObj.Args) == 0 {
			return fmt.Errorf("_min: missing column name")
		}
		col := aggrObj.Args[0]
		for _, r := range t {
			if f, ok := asFloat64(r[col]); ok {
				if !minSet || f < minVal {
					minVal = f
					minSet = true
				}
			}
		}
	default:
		return fmt.Errorf("_min: unsupported input %T", data)
	}

	if !minSet {
		return fmt.Errorf("_min: no numeric values found")
	}
	e.SetMemoryValue(stmt.Name, minVal)
	return nil
}

func _max(stmt *parser.Statement, data interface{}, aggrObj parser.Aggr, e *Evaluator) error {
	maxSet := false
	maxVal := 0.0

	switch t := data.(type) {
	case []float64:
		for _, v := range t {
			if !maxSet || v > maxVal {
				maxVal = v
				maxSet = true
			}
		}
	case []map[string]interface{}:
		if len(aggrObj.Args) == 0 {
			return fmt.Errorf("_max: missing column name")
		}
		col := aggrObj.Args[0]
		for _, r := range t {
			if f, ok := asFloat64(r[col]); ok {
				if !maxSet || f > maxVal {
					maxVal = f
					maxSet = true
				}
			}
		}
	default:
		return fmt.Errorf("_max: unsupported input %T", data)
	}

	if !maxSet {
		return fmt.Errorf("_max: no numeric values found")
	}
	e.SetMemoryValue(stmt.Name, maxVal)
	return nil
}

// _date: reduces a LIST of epoch times to a single formatted string.
// - takes the first element;
// - supports seconds or milliseconds epoch;
// - optional arg[0] can specify format (Go time layout). Default: "2006-01-02 15:04:05".
func _date(stmt *parser.Statement, data interface{}, aggrObj parser.Aggr, e *Evaluator) error {
	var layout string
	if len(aggrObj.Args) > 0 && aggrObj.Args[0] != "" {
		layout = aggrObj.Args[0]
	} else {
		layout = "2006-01-02 15:04:05"
	}

	var epoch float64
	switch t := data.(type) {
	case []float64:
		if len(t) == 0 {
			return fmt.Errorf("_date: empty list")
		}
		epoch = t[0]
	default:
		return fmt.Errorf("_date: expected LIST of numbers (epoch), got %T", data)
	}

	// seconds vs milliseconds heuristic
	sec := int64(epoch)
	if epoch > 1e12 { // looks like ms
		sec = int64(epoch / 1000)
	}
	dt := time.Unix(sec, 0).UTC()
	e.SetMemoryValue(stmt.Name, dt.Format(layout))
	return nil
}

func _unique(stmt *parser.Statement, data interface{}, aggrObj parser.Aggr, e *Evaluator) error {
	switch t := data.(type) {

	// ---------- LIST (strings) ----------
	case []string:
		seen := make(map[string]struct{}, len(t))
		out := make([]string, 0, len(t))
		for _, s := range t {
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
		e.SetMemoryValue(stmt.Name, out)
		return nil

	// ---------- LIST (numbers) ----------
	case []float64:
		seen := make(map[float64]struct{}, len(t))
		out := make([]float64, 0, len(t))
		for _, f := range t {
			if _, ok := seen[f]; ok {
				continue
			}
			seen[f] = struct{}{}
			out = append(out, f)
		}
		e.SetMemoryValue(stmt.Name, out)
		return nil

	// ---------- TABLE ----------
	case []map[string]interface{}:
		if len(aggrObj.Args) == 0 {
			return fmt.Errorf("_distinct: missing column name(s)")
		}
		cols := make([]string, len(aggrObj.Args))
		for i, a := range aggrObj.Args {
			if a == "" {
				return fmt.Errorf("_distinct: key %d must be non-empty string", i)
			}
			cols[i] = a
		}

		seen := make(map[string]struct{}, len(t))
		out := make([]map[string]interface{}, 0, len(t))

		for _, r := range t {
			key := makeCompositeKey(r, cols)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			// keep the FULL row (all columns), not just the key columns
			out = append(out, r)
		}

		e.SetMemoryValue(stmt.Name, out)
		return nil

	default:
		return fmt.Errorf("_distinct: expected LIST ([]string/[]float64) or TABLE ([]map[string]interface{}), got %T", data)
	}
}

// helper: composite key builder for DISTINCT over multiple columns (left→right)
func makeCompositeKey(row map[string]interface{}, cols []string) string {
	var b strings.Builder
	const sep = '\x1f' // unit separator
	for i, c := range cols {
		if i > 0 {
			b.WriteByte(byte(sep))
		}
		v := row[c]
		switch {
		case v == nil:
			b.WriteString("N:")
		default:
			if f, ok := asFloat64(v); ok {
				b.WriteString("F:")
				b.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
			} else {
				b.WriteString("S:")
				b.WriteString(fmt.Sprint(v))
			}
		}
	}
	return b.String()
}
