package database

import (
	"onql/storemanager"
)

// hold full schema in memory vor validations and etc
var FullSchema map[string]map[string]map[string]map[string]string
var FullSchemaTypes map[string]map[string]map[string]string
var FullSchemaStorages map[string]map[string]map[string]string

func init() {
	storemanager.SetupStore()
	SetupProtocols()
	var err error
	FullSchema, err = GetFullSchema()
	if err != nil {
		// Handle error
		panic("Failed to get full schema: " + err.Error())
	}
	SetFullSchema(FullSchema)
}

func Destruct() {
	storemanager.Destruct()
}

func SetFullSchema(schema map[string]map[string]map[string]map[string]string) {
	FullSchema = schema
	FullSchemaTypes = make(map[string]map[string]map[string]string)
	FullSchemaStorages = make(map[string]map[string]map[string]string)

	for db, tables := range schema {
		if _, ok := FullSchemaTypes[db]; !ok {
			FullSchemaTypes[db] = make(map[string]map[string]string)
		}
		if _, ok := FullSchemaStorages[db]; !ok {
			FullSchemaStorages[db] = make(map[string]map[string]string)
		}

		for table, columns := range tables {
			if _, ok := FullSchemaTypes[db][table]; !ok {
				FullSchemaTypes[db][table] = make(map[string]string)
			}
			if _, ok := FullSchemaStorages[db][table]; !ok {
				FullSchemaStorages[db][table] = make(map[string]string)
			}
			for col, attr := range columns {
				FullSchemaTypes[db][table][col] = attr["type"]
				FullSchemaStorages[db][table][col] = attr["storage"]
			}
		}
	}
	//set in storemanager also will change it later
	storemanager.SetSchema(FullSchemaStorages, FullSchemaTypes, FullSchema)
	SetProtocolBySchema("default", schema)
}
