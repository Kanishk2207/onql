package storemanager

import (
	"encoding/json"
	"fmt"
	"onql/config"
	"onql/engine"
	"os"
	"path/filepath"
	"strings"
)

//create database
//create table schema
//rename table
//rename database
//drop table
//drop database
//alter table
//get table schema
//get tables by database
//get databases

const rawFilePath = "system/schema.db"

// CreateDatabase creates a new database with the given name.
// It stores the database name in a disk store with the key "db:<name>".
// we cant validate in this file
func CreateDatabase(name string) error {
	diskStore, err := GetDiskStore(rawFilePath)
	if err != nil {
		return err
	}
	diskStore.Set("db:"+name, []byte(name))
	return nil
}

// CreateTable creates a new table with the given schema in the specified database.
func CreateTable(db, table string, schema map[string]map[string]string) error {
	diskStore, err := GetDiskStore(rawFilePath)
	if err != nil {
		return err
	}
	tableKey := "schema:" + db + ":" + table
	schemaJson, _ := json.Marshal(schema)
	err = diskStore.Set(tableKey, schemaJson)
	if err != nil {
		return err
	}
	return nil
}

// GetTableSchema retrieves the schema of a specified table in a database.
func GetTableSchema(db, table string) (map[string]map[string]string, error) {
	diskStore, err := GetDiskStore(rawFilePath)
	if err != nil {
		return nil, err
	}
	tableKey := "schema:" + db + ":" + table
	raw, err := diskStore.Get(tableKey)
	if err != nil {
		return nil, err
	}
	var schema map[string]map[string]string
	err = json.Unmarshal([]byte(raw), &schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

// GetDatabases retrieves a list of all databases.
func GetDatabases() ([]string, error) {
	diskStore, err := GetDiskStore(rawFilePath)
	if err != nil {
		return nil, err
	}
	data, err := diskStore.Scan("db:")
	if err != nil {
		return nil, err
	}
	databases := make([]string, len(data))
	for i, key := range data {
		databases[i] = strings.Split(key, ":")[1]
	}
	return databases, nil
}

// GetTables retrieves a list of all tables in a specified database.
func GetTables(db string) ([]string, error) {
	diskStore, err := GetDiskStore(rawFilePath)
	if err != nil {
		return nil, err
	}
	data, err := diskStore.Scan("schema:" + db + ":")
	if err != nil {
		return nil, err
	}
	tables := make([]string, len(data))
	for i, key := range data {
		tables[i] = strings.Split(key, ":")[2]
	}
	return tables, nil
}

// AlterTable updates the schema of a specified table in a database.
// check type of change :- renameColumn, addColumn, removeColumn, changeDataType, changeStrorageType
// get current schema - new schema
func AlterTable(db, table string, alters map[string]map[string]string) error {
	diskStore, err := GetDiskStore(rawFilePath)
	if err != nil {
		return err
	}
	tableKey := "schema:" + db + ":" + table
	raw, err := diskStore.Get(tableKey)
	if err != nil {
		return err
	}
	//load and save schema
	var schema map[string]map[string]string
	var oldSchema map[string]map[string]string
	json.Unmarshal([]byte(raw), &schema)
	json.Unmarshal([]byte(raw), &oldSchema)
	newName := ""
	oldName := ""
	//start operations
	var da DataAlter
	for action, details := range alters {
		switch action {
		case "changeBlank", "changeDefault":
			colName := details["name"]
			if colSchema, exists := schema[colName]; exists {
				if details["newDefault"] != "" {
					colSchema["default"] = details["newDefault"]
				}
				if details["newBlank"] != "" {
					colSchema["blank"] = details["newBlank"]
				}
				schema[colName] = colSchema
			}
		case "renameColumn":
			oldName = details["oldName"]
			newName = details["newName"]
			if colSchema, exists := schema[oldName]; exists {
				schema[newName] = colSchema
				delete(schema, oldName)
			}
		case "addColumn":
			newName = details["name"]
			schema[newName] = map[string]string{
				"type":    details["type"],
				"storage": details["storage"],
				"blank":   details["blank"],
				"default": details["default"],
			}
		case "dropColumn":
			delete(schema, details["name"])
		// case "changeDataType":
		// 	colName := details["name"]
		// 	if colSchema, exists := schema[colName]; exists {
		// 		colSchema["type"] = details["newType"]
		// 		schema[colName] = colSchema
		// 	}
		// case "changeStorageType":
		// 	colName := details["name"]
		// 	if colSchema, exists := schema[colName]; exists {
		// 		colSchema["storage"] = details["newStorage"]
		// 		schema[colName] = colSchema
		// 	}
		default:
			return fmt.Errorf("unknown alter action: %s", action)
		}
		// save schema here so alter data or indexes get latest
		newSchemaJson, _ := json.Marshal(schema)
		err = diskStore.Set(tableKey, newSchemaJson)
		if err != nil {
			return err
		}
		//now change data first
		da = DataAlter{
			db:        db,
			table:     table,
			oldCol:    oldName,
			newCol:    newName,
			action:    action,
			oldSchema: oldSchema,
			newSchema: schema,
		}
		alterData(da)
	}

	// regenerate indexes here
	alterIndex(da)
	return nil
}

func RenameSchemaKeys(db string, mapping map[string]string) error {
	diskStore, err := GetDiskStore(rawFilePath)
	if err != nil {
		return err
	}

	tables, err := GetTables(db)
	if err != nil {
		return err
	}

	newDB := mapping[db+"_db"]
	if newDB == "" {
		newDB = db
	}

	for _, table := range tables {
		schema, err := GetTableSchema(db, table)
		if err != nil {
			return err
		}

		newTable := mapping[table+"_table"]
		if newTable == "" {
			newTable = table
		}

		oldKey := "schema:" + db + ":" + table
		newKey := "schema:" + newDB + ":" + newTable
		if newKey == oldKey {
			continue
		}

		b, err := json.Marshal(schema) // <-- marshal to []byte
		if err != nil {
			return err
		}

		if err := diskStore.Set(newKey, b); err != nil {
			return err
		}
		// remove old key so it truly "renames"
		if err := diskStore.Delete(oldKey); err != nil {
			return err
		}
	}

	// If DB name changed, update db list entry
	if newDB != db {
		if err := diskStore.Set("db:"+newDB, []byte(newDB)); err != nil {
			return err
		}
		if err := diskStore.Delete("db:" + db); err != nil {
			return err
		}
	}

	return nil
}

func RenameDatabase(oldName, newName string) error {
	//alter data
	if err := AlterDataKeys(oldName, []string{}, map[string]string{oldName + "_db": newName}); err != nil {
		return err
	}
	//alter indexes
	if err := AlterIndexKeys(oldName, []string{}, map[string]string{oldName + "_db": newName}); err != nil {
		return err
	}
	//alter db and tables keys
	if err := RenameSchemaKeys(oldName, map[string]string{oldName + "_db": newName}); err != nil {
		return err
	}
	//rename folder
	storePath := config.Env("DISK_PATH")
	if err := RenameFolder(storePath+"/"+oldName, storePath+"/"+newName, oldName); err != nil {
		return err
	}
	return nil
}

// func RenameTable()
func RenameTable(db, oldTable, newTable string) error {
	//alter data
	if err := AlterDataKeys(db, []string{oldTable}, map[string]string{oldTable + "_table": newTable}); err != nil {
		return err
	}
	//alter indexes
	if err := AlterIndexKeys(db, []string{oldTable}, map[string]string{oldTable + "_table": newTable}); err != nil {
		return err
	}
	//alter db and tables keys
	if err := RenameSchemaKeys(db, map[string]string{oldTable + "_table": newTable}); err != nil {
		return err
	}
	//rename folder
	storePath := config.Env("DISK_PATH") + "/" + db + "/"
	if err := RenameFolder(storePath+"/"+oldTable, storePath+"/"+newTable, db); err != nil {
		return err
	}
	return nil
}

// delete db
func DeleteDatabase(db string) error {
	//delete tables
	tables, err := GetTables(db)
	if err != nil {
		return err
	}
	for _, table := range tables {
		if err := DeleteTable(db, table); err != nil {
			return err
		}
	}
	//delete db entry
	diskStore, err := GetDiskStore(rawFilePath)
	if err != nil {
		return err
	}
	if err := diskStore.Delete("db:" + db); err != nil {
		return err
	}
	err = RemoveFolder(db, "")
	if err != nil {
		return err
	}
	return nil
}

// delete table
func DeleteTable(db, table string) error {
	//delete table entry
	diskStore, err := GetDiskStore(rawFilePath)
	if err != nil {
		return err
	}
	if err := diskStore.Delete("schema:" + db + ":" + table); err != nil {
		return err
	}
	err = RemoveFolder(db, table)
	if err != nil {
		return err
	}
	return nil
}

// RenameFolder closes the disk store (to release locks) and renames/moves a folder.
// It creates the destination parent directory if needed.
func RenameFolder(oldPath, newPath string, dbname string) error {
	// 1) Close any open handles before renaming
	ds := getDbStores(dbname)
	for _, store := range ds {
		store.Close()
	}

	// 2) Sanity checks and prepare destination
	if _, err := os.Stat(oldPath); err != nil {
		return fmt.Errorf("source not accessible: %w", err)
	}
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("destination already exists: %s", newPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check destination failed: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
		return fmt.Errorf("make dest parent: %w", err)
	}

	// 3) Rename (same filesystem)
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("rename %s -> %s: %w", oldPath, newPath, err)
	}
	return nil
}

func RemoveFolder(db string, table string) error {
	folderPath := ""
	stores := make(map[string]*engine.BadgerDB)
	if table == "" {
		stores = getDbStores(db)
		folderPath = config.Env("DISK_PATH") + "/" + db
	} else {
		for path, store := range getTableDataStores(db, table) {
			stores[path] = store
		}
		for path, store := range getTableIndexStores(db, table) {
			stores[path] = store
		}
		folderPath = config.Env("DISK_PATH") + "/" + db + "/" + table
	}

	for _, store := range stores {
		store.Close()
	}

	// 2) Remove the folder
	if err := os.RemoveAll(folderPath); err != nil {
		return fmt.Errorf("remove folder %s: %w", folderPath, err)
	}
	return nil
}
