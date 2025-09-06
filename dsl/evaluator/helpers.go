package evaluator

import (
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
