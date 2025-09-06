package storemanager

import (
	"encoding/json"
	"strings"
)

const protocolFilePath = "system/protocol.db"

// QueryProtocol maps module names to their metadata definition
type QueryProtocol map[string]*Module

// Module represents a logical module, typically one database
type Module struct {
	Database string             `json:"database"` // Physical database name
	Entities map[string]*Entity `json:"entities"` // Logical tables (entities) in this module
}

// Entity represents a table with its fields, relations, and metadata
type Entity struct {
	Table     string               `json:"table"`               // Physical table name
	Fields    map[string]string    `json:"fields"`              // Field name to type mapping
	Relations map[string]*Relation `json:"relations,omitempty"` // Related entities (foreign keys)
	Context   map[string]string    `json:"context,omitempty"`   // Optional context metadata
	// Exporter  map[string]string    `json:"exporter,omitempty"`  // Export configurations
}

// Relation defines a relationship between entities
type Relation struct {
	ProtoTable string `json:"prototable"`
	Type       string `json:"type"`              // Relation type: "otm", "mtm", "mto"
	Entity     string `json:"entity"`            // Target entity name
	Through    string `json:"through,omitempty"` // Optional through entity for many-to-many
	FKField    string `json:"fk_field"`          // Format: "local:foreign"
}

// SetProtocol saves a new protocol with the given password as key.
func SetProtocol(protocol QueryProtocol, password string) error {
	diskStore, err := GetDiskStore(protocolFilePath)
	if err != nil {
		return err
	}
	key := "protocol:" + password
	data, err := json.Marshal(protocol)
	if err != nil {
		return err
	}
	return diskStore.Set(key, data)
}

func GetProtocolByPassword(password string) (QueryProtocol, error) {
	diskStore, err := GetDiskStore(protocolFilePath)
	if err != nil {
		return nil, err
	}
	raw, err := diskStore.Get("protocol:" + password)
	if err != nil {
		return nil, err
	}
	var protocols QueryProtocol
	if err := json.Unmarshal([]byte(raw), &protocols); err != nil {
		return nil, err
	}
	return protocols, nil
}

// // GetProtocol returns the saved protocol for the given password.
// func GetProtocol(password string) (QueryProtocol, error) {
// 	diskStore, err := GetDiskStore(protocolFilePath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	raw, err := diskStore.Get("protocol:" + password)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var protocols QueryProtocol
// 	if err := json.Unmarshal([]byte(raw), &protocols); err != nil {
// 		return nil, err
// 	}
// 	return protocols, nil
// }

// GetProtocols returns all saved protocols.
func GetProtocols() ([]QueryProtocol, error) {
	diskStore, err := GetDiskStore(protocolFilePath)
	if err != nil {
		return nil, err
	}
	raw, err := diskStore.ScanWithValues("protocol:")
	if err != nil {
		return nil, err
	}
	var protocols []QueryProtocol
	for _, item := range raw {
		var protocol QueryProtocol
		if err := json.Unmarshal([]byte(item), &protocol); err != nil {
			return nil, err
		}
		protocols = append(protocols, protocol)
	}
	return protocols, nil
}

// DeleteProtocol removes a protocol by password.
func DeleteProtocol(password string) error {
	diskStore, err := GetDiskStore(protocolFilePath)
	if err != nil {
		return err
	}
	key := "protocol:" + password
	return diskStore.Delete(key)
}

// GetProtocolIDs returns all stored protocol identifiers (password keys)
func GetProtocolIDs() ([]string, error) {
	diskStore, err := GetDiskStore(protocolFilePath)
	if err != nil {
		return nil, err
	}
	keys, err := diskStore.Scan("protocol:")
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(keys))
	for i, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) > 1 {
			ids[i] = parts[1]
		}
	}
	return ids, nil
}

/*

func (q *QueryProtocol) IsColumn(value string) bool {
	for _, module := range *q {
		for _, entity := range module.Entities {
			if _, exists := entity.Fields[value]; exists {
				return true
			}
		}
	}
	return false
}

func (q *QueryProtocol) IsDatabase(value string) bool {
	for _, module := range *q {
		if module.Database == value {
			return true
		}
	}
	return false
}

func (q *QueryProtocol) IsTable(value string) bool {
	for _, module := range *q {
		for _, entity := range module.Entities {
			if entity.Table == value {
				return true
			}
		}
	}
	return false
}

*/
