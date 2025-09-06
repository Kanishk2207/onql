package database

import (
	"errors"
	"onql/storemanager"
)

// Delete deletes one or more records from the specified table.
// It requires a slice of primary key strings (ids) to identify which records to delete.
// Example usage:
// err := database.Delete("mydb", "mytable", []string{"pk1", "pk2"})
func Delete(db, table string, ids []string) error {
	// Check if database exists
	if !IsDatabaseExists(db) {
		return errors.New("database does not exist")
	}

	// Check if table exists
	if !IsTableExists(db, table) {
		return errors.New("table does not exist")
	}

	// Check if at least one id is provided
	if len(ids) == 0 {
		return errors.New("no primary keys provided for deletion")
	}

	// Perform deletion via storemanager
	if err := storemanager.DeleteData(db, table, ids); err != nil {
		return err
	}

	return nil
}
