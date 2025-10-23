package evaluator

import (
	"errors"
	"fmt"
	"onql/dsl/parser"
	"strconv"
	"strings"
)

func (e *Evaluator) EvalFilter() error {
	// Implement filtering logic here
	filterStmt := e.Plan.NextStatement(true)
	if filterStmt.Operation != parser.OpStartFilter {
		return errors.New("expected start filter operation")
	}
	result := make([]map[string]interface{}, 0)
	// stmt := e.Plan.NextStatement(true)
	pos := e.Plan.Pos
	var endFilterPos int
	var endFilterName string
	// Continue with filtering logic
	// fmt.Println(e.Memory)
	// fmt.Println(filterStmt.Sources[0].SourceValue)
	//get data from start filter table
	tableData, ok := e.Memory[filterStmt.Sources[0].SourceValue].([]map[string]interface{})
	if !ok {
		return errors.New("expect table data for filter but got " + fmt.Sprintf("%T", e.Memory[filterStmt.Sources[0].SourceValue]))
	}
	if len(tableData) == 0 {
		nested := 0
		for {
			stmt := e.Plan.NextStatement(true)
			if stmt == nil {
				return fmt.Errorf("expect ] but got empty in projection")
			}
			if stmt.Operation == parser.OpEndFilter {
				if nested > 0 {
					nested -= 1
					continue
				}
				endFilterPos = e.Plan.Pos
				endFilterName = stmt.Name
				break
			}
			if stmt.Operation == parser.OpStartFilter {
				nested += 1
			}
		}
	} else {
		for _, row := range tableData {
			// Apply filter conditions
			// e.Memory[filterStmt.Name] = row
			e.SetMemoryValue(filterStmt.Name, row)
			for {
				stmt := e.Plan.NextStatement(false)
				if stmt == nil || stmt.Operation == parser.OpEndFilter {
					endFilterPos = e.Plan.Pos
					endFilterName = stmt.Name
					e.Plan.NextStatement(true) // move to next statement
					break
				}
				err := e.EvalStatement()
				if err != nil {
					return err
				}
			}
			prevStmt := e.Plan.PrevStatement(false)
			conditionResult := e.Memory[prevStmt.Name].(bool)
			if conditionResult {
				result = append(result, row)
			}
			// Restore original position for next filter
			e.Plan.Pos = pos
		}
	}
	e.Plan.Pos = endFilterPos
	e.Plan.NextStatement(true) // Move past OpEndFilter
	// e.Memory[endFilterName] = result
	e.SetMemoryValue(endFilterName, result)
	fmt.Println(result)
	return nil
}

func (e *Evaluator) isUnderFilter(varName string) bool {

	stmt := e.Plan.StatementMap[varName]
	if stmt.Operation == parser.OpAccessRelatedTable {
		stmt = e.Plan.StatementMap[e.Plan.StatementMap[varName].Sources[1].SourceValue]
	} else {
		stmt = e.Plan.StatementMap[e.Plan.StatementMap[varName].Sources[0].SourceValue]
	}
	return stmt.Operation == parser.OpStartFilter
}

func (e *Evaluator) IsUnderProjection(varName string) bool {
	// stmt := e.Plan.StatementMap[e.Plan.StatementMap[varName].Sources[0].SourceValue]
	stmt := e.Plan.StatementMap[varName]
	if stmt.Operation == parser.OpAccessRelatedTable {
		stmt = e.Plan.StatementMap[e.Plan.StatementMap[varName].Sources[1].SourceValue]
	} else {
		stmt = e.Plan.StatementMap[e.Plan.StatementMap[varName].Sources[0].SourceValue]
	}
	return stmt.Operation == parser.OpStartProjectionKey || stmt.Operation == parser.OpStartProjection
}

func (e *Evaluator) EvalSlice() error {
	stmt := e.Plan.NextStatement(true)
	if stmt.Operation != parser.OpSlice {
		return fmt.Errorf("expected slice operation")
	}
	sliceParts := strings.Split(stmt.Expressions.(string), ":")
	data := e.Memory[stmt.Sources[0].SourceValue]

	var arr []interface{}

	switch s := data.(type) {
	case []map[string]interface{}:
		arr = make([]interface{}, len(s))
		for i, v := range s {
			arr[i] = v
		}
	case []string:
		arr = make([]interface{}, len(s))
		for i, v := range s {
			arr[i] = v
		}
	case []float64:
		arr = make([]interface{}, len(s))
		for i, v := range s {
			arr[i] = v
		}
	case []int64:
		arr = make([]interface{}, len(s))
		for i, v := range s {
			arr[i] = v // or float64(v)
		}
	case []interface{}:
		arr = s
	default:
		return errors.New("expected array for row access for table row access")
	}

	arrLen := len(arr)

	// Parse step first to determine direction
	step := 1
	var err error
	if len(sliceParts) > 2 && sliceParts[2] != "" {
		step, err = strconv.Atoi(sliceParts[2])
		if err != nil {
			return fmt.Errorf("invalid step: %v", err)
		}
		if step == 0 {
			return fmt.Errorf("slice step cannot be zero")
		}
	}

	// Use pointers to track whether values were explicitly provided
	var startPtr, endPtr *int

	// Parse start if provided
	if len(sliceParts) > 0 && sliceParts[0] != "" {
		s, err := strconv.Atoi(sliceParts[0])
		if err != nil {
			return fmt.Errorf("invalid start index: %v", err)
		}
		startPtr = &s
	}

	// Parse end if provided
	if len(sliceParts) > 1 && sliceParts[1] != "" {
		e, err := strconv.Atoi(sliceParts[1])
		if err != nil {
			return fmt.Errorf("invalid end index: %v", err)
		}
		endPtr = &e
	}

	// Normalize indices and apply defaults based on step direction
	var actualStart, actualEnd int

	if step > 0 {
		// Forward slicing defaults and normalization
		if startPtr == nil {
			actualStart = 0
		} else {
			actualStart = *startPtr
			// Handle negative indices
			if actualStart < 0 {
				actualStart += arrLen
			}
			// Clamp to valid range
			if actualStart < 0 {
				actualStart = 0
			} else if actualStart > arrLen {
				actualStart = arrLen
			}
		}

		if endPtr == nil {
			actualEnd = arrLen
		} else {
			actualEnd = *endPtr
			// Handle negative indices
			if actualEnd < 0 {
				actualEnd += arrLen
			}
			// Clamp to valid range
			if actualEnd < 0 {
				actualEnd = 0
			} else if actualEnd > arrLen {
				actualEnd = arrLen
			}
		}
	} else {
		// Backward slicing defaults and normalization
		if startPtr == nil {
			actualStart = arrLen - 1
		} else {
			actualStart = *startPtr
			// Handle negative indices
			if actualStart < 0 {
				actualStart += arrLen
			}
			// Clamp to valid range
			if actualStart >= arrLen {
				actualStart = arrLen - 1
			}
			// If still negative after normalization, the slice will be empty
			if actualStart < 0 {
				e.SetMemoryValue(stmt.Name, []interface{}{})
				return nil
			}
		}

		if endPtr == nil {
			// Default end for negative step is "before first element"
			actualEnd = -1
		} else {
			actualEnd = *endPtr
			// Handle negative indices
			if actualEnd < 0 {
				actualEnd += arrLen
			}
			// For negative step, actualEnd is exclusive (we stop at actualEnd + 1)
			// Clamp to valid range
			if actualEnd < -1 {
				actualEnd = -1
			} else if actualEnd >= arrLen {
				actualEnd = arrLen - 1
			}
		}
	}

	// Build result
	result := make([]interface{}, 0)

	if step > 0 {
		// Forward iteration: go from actualStart to actualEnd (exclusive)
		for i := actualStart; i < actualEnd; i += step {
			result = append(result, arr[i])
		}
	} else {
		// Backward iteration: go from actualStart down to actualEnd (exclusive)
		for i := actualStart; i > actualEnd; i += step {
			result = append(result, arr[i])
		}
	}

	e.SetMemoryValue(stmt.Name, result)
	return nil
}
