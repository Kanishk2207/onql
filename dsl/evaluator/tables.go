package evaluator

import (
	"errors"
	"fmt"
	"onql/database"
	"onql/dsl/parser"
	"onql/storemanager"
	"strconv"
	"strings"
)

func (e *Evaluator) EvalTableWithContext() error {
	// Implement table evaluation logic here
	stmt := e.Plan.NextStatement(true)
	if stmt.Operation != parser.OpAccessTable {
		return errors.New("expected access table operation")
	}
	// if e.ContextKey == "" {
	// 	e.Plan.PrevStatement(true) //move back
	// 	return e.EvalTable()
	// }
	sources := strings.Split(stmt.Sources[0].SourceValue, ".")
	cntxQuery, err := database.GetProtoContext(e.Plan.ProtocolPass, sources[0], sources[1], e.ContextKey)
	if err != nil {
		return err
	}
	if cntxQuery == "" {
		// e.Plan.PrevStatement(true) //move back
		//moveback forcefully
		e.Plan.Pos = e.Plan.Pos - 1
		return e.EvalTable()
	}
	for i, v := range e.ContextValues {
		cntxQuery = strings.Replace(cntxQuery, "$"+strconv.Itoa(i+1), "\""+v+"\"", 1)
	}
	lexer := parser.NewLexer(cntxQuery)
	plan := parser.NewPlan(lexer, e.Plan.ProtocolPass)
	err = plan.Parse()
	if err != nil {
		return err
	}
	eval := NewEvaluator(plan, "", []string{cntxQuery})
	err = eval.Eval()
	if err != nil {
		return err
	}
	e.SetMemoryValue(stmt.Name, eval.Result)
	// e.Memory[stmt.Name] = eval.Result
	// e.Memory[stmt.Name+"_meta_structure_type"] = getStructureType(eval.Result)
	return nil
}

func (e *Evaluator) EvalTable() error {
	// Implement table evaluation logic here
	stmt := e.Plan.NextStatement(true)
	if stmt.Operation != parser.OpAccessTable {
		return errors.New("expected access table operation")
	}
	// Continue with table evaluation logic
	data, err := GetTableData(stmt.Meta["db"], stmt.Meta["table"])
	if err != nil {
		return err
	}
	// e.Memory[stmt.Name] = data
	e.SetMemoryValue(stmt.Name, data)
	// fmt.Println(data)
	// e.Memory[stmt.Name] = data
	return nil
}

func (e *Evaluator) EvalRelatedTable() error {
	// Implement related table evaluation logic here
	stmt := e.Plan.NextStatement(true)
	if stmt.Operation != parser.OpAccessRelatedTable {
		return errors.New("expected access table operation")
	}

	fkKey := strings.Split(stmt.Expressions.(*storemanager.Relation).FKField, ":")[0]
	result := make([]map[string]interface{}, 0)
	if e.isUnderFilter(stmt.Name) || e.IsUnderProjection(stmt.Name) {
		val, ok := e.Memory[stmt.Sources[1].SourceValue].(map[string]interface{})
		if !ok {
			return fmt.Errorf("host table data not found for geting related table data %s")
		}
		data, err := GetRelatedTableData(stmt.Meta["db"], *stmt.Expressions.(*storemanager.Relation), val[fkKey].(string))
		if err != nil {
			return err
		}
		result = append(result, data...)
	} else {
		tabledata, ok := e.Memory[stmt.Sources[1].SourceValue].([]map[string]interface{})
		if !ok {
			return fmt.Errorf("host table data not found for geting related table data %s")
		}
		for _, val := range tabledata {
			data, err := GetRelatedTableData(stmt.Meta["db"], *stmt.Expressions.(*storemanager.Relation), val[fkKey].(string))
			if err != nil {
				return err
			}
			result = append(result, data...)
		}
	}

	// e.Memory[stmt.Name] = data
	e.SetMemoryValue(stmt.Name, result)
	return nil
}

func (e *Evaluator) EvalTableList() error {
	// Implement table list evaluation logic here
	stmt := e.Plan.NextStatement(true)
	if stmt.Operation != parser.OpAccessList {
		return errors.New("expected access table list operation")
	}
	if e.isUnderFilter(stmt.Name) || e.IsUnderProjection(stmt.Name) {
		e.Plan.PrevStatement(true)
		return e.EvalTableField()
	}
	sourceName := stmt.Sources[0].SourceValue
	list := make([]string, 0)
	listNum := make([]float64, 0)
	listOther := make([]interface{}, 0)
	stmtMetadataType := strings.ToUpper(stmt.Meta["type"])
	for _, item := range e.Memory[sourceName].([]map[string]interface{}) {
		if val, ok := item[stmt.Meta["name"]]; ok {
			if stmtMetadataType == "NUMBER" || stmtMetadataType == "TIMESTAMP" {
				// num, err := strconv.ParseFloat(val.(string), 64)
				// if err != nil {
				// return err
				// }
				num := val.(float64)
				listNum = append(listNum, num)
			} else if stmtMetadataType == "STRING" {
				list = append(list, val.(string))
			} else {
				listOther = append(listOther, val)
			}
		}
	}
	if stmtMetadataType == "NUMBER" || stmtMetadataType == "TIMESTAMP" {
		// e.Memory[stmt.Name] = listNum
		e.SetMemoryValue(stmt.Name, listNum)
	} else if stmtMetadataType == "STRING" {
		// e.Memory[stmt.Name] = list
		e.SetMemoryValue(stmt.Name, list)
	} else {
		e.SetMemoryValue(stmt.Name, listOther)
	}
	// e.Memory[stmt.Name+"_meta_structure_type"] = "LIST"
	// e.Memory[stmt.Name+"_meta_type"] = stmt.Meta["type"]
	return nil
}

func (e *Evaluator) EvalTableRow() error {
	// Implement table row evaluation logic here
	stmt := e.Plan.NextStatement(true)
	if stmt.Operation != parser.OpAccessRow {
		return errors.New("expected access table row operation")
	}
	sourceName := stmt.Sources[0].SourceValue
	// row := make(map[string]interface{})
	// if e.Plan.StatementMap[stmt.Expressions.(string)].Operation == parser.OpUnknownIdentifier {

	// } else {

	// }
	switch e.Memory[sourceName].(type) {
	case []map[string]interface{}:
		row2 := e.Memory[sourceName].([]map[string]interface{})[stmt.Expressions.(int64)]
		// e.Memory[stmt.Name] = row2
		e.SetMemoryValue(stmt.Name, row2)
	case []interface{}:
		row2 := e.Memory[sourceName].([]interface{})[stmt.Expressions.(int64)]
		e.SetMemoryValue(stmt.Name, row2)
	case []string:
		e.SetMemoryValue(stmt.Name, e.Memory[sourceName].([]string)[stmt.Expressions.(int64)])
		// e.Memory[stmt.Name+"_meta_type"] = "STRING"
	case []float64, []int64:
		e.SetMemoryValue(stmt.Name, e.Memory[sourceName].([]float64)[stmt.Expressions.(int64)])
		// e.Memory[stmt.Name+"_meta_type"] = "NUMBER"
	default:
		return errors.New("expected array for row access for table row access")
	}
	return nil
}

func (e *Evaluator) EvalTableField() error {
	// Implement table field evaluation logic here
	stmt := e.Plan.NextStatement(true)
	if stmt.Operation != parser.OpAccessField && (stmt.Operation != parser.OpAccessList || (!e.isUnderFilter(stmt.Name) && !e.IsUnderProjection(stmt.Name))) {
		return errors.New("expected access table field operation")
	}
	sourceName := stmt.Sources[0].SourceValue

	field := e.Memory[sourceName].(map[string]interface{})[stmt.Meta["name"]]
	// e.Memory[stmt.Name] = field
	// e.Memory[stmt.Name+"_meta_type"] = stmt.Meta["type"]
	e.SetMemoryValue(stmt.Name, field)
	return nil
}

func (e *Evaluator) EvalLiteral() error {
	// Implement literal evaluation logic here
	stmt := e.Plan.NextStatement(true)
	if stmt.Operation != parser.OpLiteral {
		return errors.New("expected literal operation")
	}
	if stmt.Meta["type"] == "NUMBER" {
		num, err := strconv.ParseFloat(stmt.Expressions.(string), 64)
		if err != nil {
			return err
		}
		// e.Memory[stmt.Name] = num
		e.SetMemoryValue(stmt.Name, num)
	} else {
		e.SetMemoryValue(stmt.Name, stmt.Expressions)
	}
	// e.Memory[stmt.Name+"_meta_type"] = stmt.Meta["type"]
	return nil
}
