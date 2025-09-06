package database

import (
	"errors"
	"fmt"
	"onql/storemanager"
)

// Protocols holds all loaded query protocols in memory, keyed by password
var Protocols map[string]storemanager.QueryProtocol

// SetProtocol validates and stores a protocol definition using a password key
func SetProtocol(password string, protocol storemanager.QueryProtocol) error {
	if password == "" {
		return errors.New("password is required")
	}
	if len(protocol) == 0 {
		return errors.New("protocol cannot be empty")
	}
	err := ValidateProtocol(protocol)
	if err != nil {
		return errors.New("failed to validate protocol: " + err.Error())
	}
	if err := storemanager.SetProtocol(protocol, password); err != nil {
		return errors.New("failed to store protocol: " + err.Error())
	}
	return SetupProtocols()
}

// GetAllProtocols returns all stored protocols from disk
func GetAllProtocols() (map[string]storemanager.QueryProtocol, error) {
	return Protocols, nil
}

// func GetProtocolBy

// DeleteProtocolByPassword deletes a protocol by its password key
func DeleteProtocolByPassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}
	if err := storemanager.DeleteProtocol(password); err != nil {
		return errors.New("failed to delete protocol: " + err.Error())
	}
	return SetupProtocols()
}

// GetProtocolPasswords returns all saved protocol password keys
func GetProtocolPasswords() ([]string, error) {
	return storemanager.GetProtocolIDs()
}

// SetupProtocols reloads all protocols from disk into memory
func SetupProtocols() error {
	// rawProtocols, err := storemanager.GetProtocols()
	// if err != nil {
	// return err
	// }
	ids, err := storemanager.GetProtocolIDs()
	if err != nil {
		return err
	}
	// if len(ids) != len(rawProtocols) {
	// return errors.New("protocol ID and data count mismatch")
	// }

	Protocols = make(map[string]storemanager.QueryProtocol)
	for _, id := range ids {
		proto, err := storemanager.GetProtocolByPassword(id)
		if err != nil {
			return err
		}
		Protocols[id] = proto
	}
	return nil
}

// IsModule checks if a protocol with the given password exists
func IsModule(password string) bool {
	_, exists := Protocols[password]
	return exists
}

// IsTable checks if an entity exists within a protocol
func IsTable(password, database string, entity string) bool {
	proto, ok := Protocols[password]
	if !ok {
		return false
	}
	if !IsDatabase(password, database) {
		return false
	}
	mod := proto[database]
	if _, entityExists := mod.Entities[entity]; entityExists {
		return true
	}
	return false
}

func IsRelatedTableByRelationName(password, database, hosttable, relationName string) bool {
	if !IsTable(password, database, hosttable) {
		return false
	}
	// 1) find the protocol
	proto, ok := Protocols[password]
	if !ok {
		return false
	}

	// 2) find the module
	mod, ok := proto[database]
	if !ok {
		return false
	}

	// 3) find the source entity by alias
	ent, ok := mod.Entities[hosttable]
	if !ok {
		return false
	}

	// 4) scan only that entity’s Relations map for a link to targetEntity
	for relName, _ := range ent.Relations {
		if relName == relationName {
			return true
		}
	}

	return false
}

func GetRelationByRelationName(password, database, hosttable, relationName string) (*storemanager.Relation, error) {
	if !IsTable(password, database, hosttable) {
		return nil, fmt.Errorf("hosttable is not exists in protocol")
	}
	// 1) find the protocol
	proto := Protocols[password]
	// if !ok {
	// 	return false
	// }

	// 2) find the module
	mod, ok := proto[database]
	if !ok {
		return nil, fmt.Errorf("database is not exists in protocol")
	}

	// 3) find the source entity by alias
	ent := mod.Entities[hosttable]
	// if !ok {
	// 	return false
	// }

	// 4) scan only that entity’s Relations map for a link to targetEntity
	for relName, rel := range ent.Relations {
		if relName == relationName {
			return rel, nil
		}
	}

	return nil, fmt.Errorf("relation name not exists in " + hosttable + "relations")
}

// IsColumn checks if a column exists in a specific entity of a protocol
func IsColumn(password, database, entity, column string) bool {
	proto, ok := Protocols[password]
	if !ok {
		return false
	}
	if !IsDatabase(password, database) {
		return false
	}
	for database_alias, mod := range proto {
		if database_alias == database {
			if ent, exists := mod.Entities[entity]; exists {
				if _, colExists := ent.Fields[column]; colExists {
					return true
				}
			}
		}
	}
	return false
}

func IsDatabase(password, database string) bool {
	proto, ok := Protocols[password]
	if !ok {
		return false
	}

	for alias, _ := range proto {
		if alias == database {
			return true
		}
	}

	return false
}

// GetProtoRelation looks at the Relations defined on sourceEntity (by alias)
// and returns the one pointing to targetEntity (by alias). It does not inspect
// the reverse side or do any join-table logic.
func GetProtoRelation(
	password, // protocol key
	dbAlias, // module alias in your protocol
	sourceEntity, // entity alias in Module.Entities
	targetEntity string, // entity alias to find
) (*storemanager.Relation, error) {
	// 1) find the protocol
	proto, ok := Protocols[password]
	if !ok {
		return nil, fmt.Errorf("protocol %q not found", password)
	}

	// 2) find the module
	mod, ok := proto[dbAlias]
	if !ok {
		return nil, fmt.Errorf("module %q not defined in protocol", dbAlias)
	}

	// 3) find the source entity by alias
	ent, ok := mod.Entities[sourceEntity]
	if !ok {
		return nil, fmt.Errorf("entity %q not found in module %q", sourceEntity, dbAlias)
	}

	// 4) scan only that entity’s Relations map for a link to targetEntity
	for _, rel := range ent.Relations {
		if rel.Entity == targetEntity {
			return rel, nil
		}
	}

	return nil, fmt.Errorf(
		"no relation defined on %q pointing to %q in module %q",
		sourceEntity, targetEntity, dbAlias,
	)
}

func GetProtoContext(password, database, table, ctxKey string) (string, error) {
	proto, ok := Protocols[password]
	if !ok {
		return "", fmt.Errorf("protocol %q not found", password)
	}

	db, ok := proto[database]
	if !ok {
		return "", fmt.Errorf("database %q not found in protocol", database)
	}

	tbl, ok := db.Entities[table]
	if !ok {
		return "", fmt.Errorf("table %q not found in database %q", table, database)
	}

	return tbl.Context[ctxKey], nil
}

// func GetColumnMetaData(password, database, table, column string) (map[string]string, error) {
// 	meta := make(map[string]string)

// 	meta["dbColumnName"] = Protocols[password][database][table].Fields[column].Name
// 	// meta["dbColumnType"] = Protocols[password][database][table].Fields[column].DataType

// 	return meta, nil
// }

func GetDbNameFromProtoName(password, database string) (string, error) {
	proto, ok := Protocols[password]
	if !ok {
		return "", fmt.Errorf("protocol %q not found", password)
	}

	db, ok := proto[database]
	if !ok {
		return "", fmt.Errorf("database %q not found in protocol", database)
	}

	return db.Database, nil
}

func GetTableNameFromProtoName(password, database, table string) (string, error) {
	proto, ok := Protocols[password]
	if !ok {
		return "", fmt.Errorf("protocol %q not found", password)
	}

	db, ok := proto[database]
	if !ok {
		return "", fmt.Errorf("database %q not found in protocol", database)
	}

	return db.Entities[table].Table, nil
}

func GetColSchemaFromProtoName(password, database, table, column string) (map[string]string, error) {
	meta := make(map[string]string)
	proto, _ := Protocols[password]
	db := proto[database]
	dbName := db.Database
	tableName := db.Entities[table].Table
	meta["name"] = db.Entities[table].Fields[column]
	meta["type"] = FullSchema[dbName][tableName][meta["name"]]["type"]
	meta["storage"] = FullSchema[dbName][tableName][meta["name"]]["storage"]
	return meta, nil
}

func SetProtocolBySchema(password string, schema map[string]map[string]map[string]map[string]string) {
	// fmt.Println(schema)
	//convert schema to Query protocol
	protocol := storemanager.QueryProtocol{}
	for db, tables := range schema {
		protocol[db] = &storemanager.Module{Database: db}
		protocol[db].Entities = make(map[string]*storemanager.Entity)
		for table, columns := range tables {
			entity := storemanager.Entity{}
			entity.Table = table
			entity.Fields = make(map[string]string)
			for column := range columns {
				entity.Fields[column] = column
			}
			protocol[db].Entities[table] = &entity
		}
	}
	// fmt.Println("===================")
	fmt.Println(protocol)
	SetProtocol(password, protocol)
}
