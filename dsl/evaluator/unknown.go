package evaluator

import (
	"fmt"
	"onql/dsl/parser"
)

func (e *Evaluator) EvalUnknown() error {
	stmt := e.Plan.NextStatement(true)
	if stmt.Operation != parser.OpUnknownIdentifier {
		return fmt.Errorf("expected unknown identifier operation")
	}
	// Continue with unknown identifier evaluation logic
	data := e.Memory[stmt.Sources[0].SourceValue]
	// fmt.Println(data)
	switch data := data.(type) {
	// Handle string case
	case map[string]interface{}:
		e.SetMemoryValue(stmt.Name, data[stmt.Expressions.(string)])
		// e.Memory[stmt.Name] = data[stmt.Expressions.(string)]
		// e.Memory[stmt.Name+"_meta_type"] = getStructureType(data[stmt.Expressions.(string)])
	case []map[string]interface{}:
		result := make([]interface{}, 0)
		for _, item := range data {
			result = append(result, item[stmt.Expressions.(string)])
		}
		e.SetMemoryValue(stmt.Name, result)
		// e.Memory[stmt.Name] = result
		// e.Memory[stmt.Name+"_meta_type"] = getStructureType(result)
	case []interface{}:
		result := make([]interface{}, 0)
		for _, item := range data {
			result = append(result, item.(map[string]interface{})[stmt.Expressions.(string)])
		}
		e.SetMemoryValue(stmt.Name, result)
		// e.Memory[stmt.Name] = result
		// e.Memory[stmt.Name+"_meta_type"] = getStructureType(result)
	// cases interface{}

	default:
		return fmt.Errorf("unsupported data type %T for unknown identifier", data)
	}
	return nil
}
