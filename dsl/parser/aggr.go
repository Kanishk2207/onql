package parser

import (
	"errors"
	"fmt"
)

var AggrRegistry = map[string]map[string]string{
	"_sum":      {"LIST": "NUMBER", "TABLE": "NUMBER"},
	"_count":    {"LIST": "NUMBER", "TABLE": "NUMBER"},
	"_avg":      {"LIST": "NUMBER"},
	"_min":      {"LIST": "NUMBER"},
	"_max":      {"LIST": "NUMBER"},
	"_date":     {"LIST": "STRING"},
	"_unique": {"LIST": "LIST", "TABLE": "TABLE"},
	"_asc":      {"LIST": "LIST", "TABLE": "TABLE"},
	"_desc":     {"LIST": "LIST", "TABLE": "TABLE"},
}

func (plan *Plan) ParseAggr(stmt *Statement, dependency string) error {
	token := plan.lexer.Next(true)
	if token.Type != TOKEN_IDENTIFIER {
		return fmt.Errorf("expect identifier but got %s", token.Value)
	}
	if _, ok := AggrRegistry[token.Value]; !ok {
		return fmt.Errorf("unknown aggregate function %s", token.Value)
	}

	returnType, err := plan.GetAggrReturnType(token.Value, dependency)
	if err != nil {
		return err
	}
	inputStmt := plan.StatementMap[dependency]
	stmt.Operation = OpAggregateReduce
	stmt.Sources[0] = NewSource("var", dependency)
	stmt.Meta = map[string]string{
		"input_type":  plan.GetAggrInputTypeFromOperationType(inputStmt.Operation),
		"return_type": returnType,
	}
	aggrObj := Aggr{Name: token.Value, Args: []string{}}
	if plan.lexer.Next(false) != nil && plan.lexer.Next(false).Type == TOKEN_LPAREN {
		plan.lexer.Next(true) // consume the (
		for {
			token = plan.lexer.Next(true)
			if token.Type == TOKEN_RPAREN {
				break
			}
			if token.Type == TOKEN_COMMA {
				continue
			}
			if token.Type != TOKEN_IDENTIFIER && token.Type != TOKEN_NUMBER && token.Type != TOKEN_STRING {
				return fmt.Errorf("expect identifier | number | string but got %s", token.Value)
			}
			aggrObj.Args = append(aggrObj.Args, token.Value)
		}
	}
	stmt.Expressions = aggrObj
	return nil
}

func (plan *Plan) GetAggrReturnType(aggrName, inputStmtName string) (string, error) {
	inputStmt := plan.StatementMap[inputStmtName]
	inpType := ""
	switch inputStmt.Operation {
	case OpAccessTable, OpAccessRelatedTable:
		inpType = "TABLE"
	case OpAccessList:
		inpType = "LIST"
	case OpAccessField:
		inpType = "FIELD"
	case OpAccessRow:
		inpType = "ROW"
	case OpAggregateReduce:
		inpType = inputStmt.Meta["return_type"]
	}
	returnType, ok := AggrRegistry[aggrName][inpType]
	if !ok {
		return "", errors.New("aggr return type not found")
	}
	return returnType, nil
}

func (plan *Plan) GetOperationTypeFromAggrReturnType(returnType string) OperationType {
	switch returnType {
	case "NUMBER", "STRING":
		return OpLiteral
	case "TABLE":
		return OpAccessTable
	case "FIELD":
		return OpAccessField
	case "ROW":
		return OpAccessRow
	case "LIST":
		return OpAccessList
	}
	return OpUnknownIdentifier
}

func (plan *Plan) GetAggrInputTypeFromOperationType(ot OperationType) string {
	switch ot {
	case OpLiteral:
		return "NUMBER"
	case OpAccessTable, OpAccessRelatedTable:
		return "TABLE"
	case OpAccessField:
		return "FIELD"
	case OpAccessRow:
		return "ROW"
	case OpAccessList:
		return "LIST"
	}
	return "unknown"
}

func (plan *Plan) IsAggr(name string) bool {
	if _, ok := AggrRegistry[name]; ok {
		return true
	}
	return false
}

// func (plan *Plan) GetAggrReturnType(stmt *Statement, aggr ) string {
// 	if aggr, ok := AggrRegistry[aggrName]; ok {
// 		return aggr["return"]
// 	}
// 	return "unknown"
// }
