package dsl

// import (
// 	"fmt"
// 	"onql/dsl/evaluator"
// 	"onql/dsl/parser"
// 	"strings"
// )

// func Test(query string) {
// 	lexer := parser.NewLexer(query)
// 	// for t := lexer.Next(true); t != nil;{
// 	// 	fmt.Println(t.Value)
// 	// 	t = lexer.Next(true)
// 	// }
// 	// for {
// 	// 	token := lexer.Next(true)
// 	// 	if token == nil {
// 	// 		return
// 	// 	}
// 	// 	fmt.Printf("Token: Type=%s, Value=%s, Pos=%d\n", token.Type, token.Value, token.Pos)
// 	// }
// 	parser := parser.NewPlan(lexer, "123")
// 	err := parser.Parse()
// 	if err != nil {
// 		fmt.Println("Error parsing statement:", err)
// 		// return
// 	}
// 	fmt.Println("Parsed Statements:")
// 	// for _, stmt := range parser.Statements {
// 	// 	fmt.Printf("Name: %s, Operation: %s, Sources: %+v, Expressions: %+v\n",
// 	// 		stmt.Name, stmt.Operation, stmt.Sources, stmt.Expressions)
// 	// }
// 	// print(len(parser.Statements))
// 	for _, stmt := range parser.Statements {
// 		fmt.Println(stmt.Name, stmt.Operation)
// 	}

// 	printStatements(parser.Statements)

// 	evaluator := evaluator.NewEvaluator(parser)
// 	err = evaluator.Eval()
// 	if err != nil {
// 		fmt.Println("Error evaluating:", err)
// 	}
// 	// fmt.Println(evaluator.Memory[""])
// 	for key, value := range evaluator.Memory {
// 		// fmt.Printf("Key: %s, Value: %v\n", key, value)
// 		fmt.Println("Key:", key, "Value:", value)
// 	}
// }

// // func printStatements(stmts []*parser.Statement) {
// // 	for _, stmt := range stmts {
// // 		print(stmt.Name, stmt.Operation, stmt.Expressions)
// // 		for _, source := range stmt.Sources {
// // 			if source.SourceType == "" {
// // 				break
// // 			}
// // 			print(source.SourceType, source.SourceValue)
// // 		}
// // 		print("\n")
// // 	}
// // }

// func printStatements(stmts []*parser.Statement) {
// 	fmt.Printf("%-8s %-8s %-30s %-20s\n", "Name", "Operation", "Sources", "Expressions")
// 	fmt.Println(strings.Repeat("-", 80))
// 	for _, stmt := range stmts {
// 		// Collect sources as a comma-separated string
// 		var sources []string
// 		for _, source := range stmt.Sources {
// 			if source.SourceType == "" {
// 				continue
// 			}
// 			sources = append(sources, fmt.Sprintf("%s:%s", source.SourceType, source.SourceValue))
// 		}
// 		// Print statement info
// 		fmt.Printf("%-8s %-8s %-30s %-20v\n",
// 			stmt.Name,
// 			stmt.Operation,
// 			strings.Join(sources, ", "),
// 			stmt.Expressions,
// 		)
// 	}
// }
