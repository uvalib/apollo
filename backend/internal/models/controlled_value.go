package models

import (
	"log"
)

// ControlledValue is a controlled vocabulary for node values
type ControlledValue struct {
	ID       int64  `json:"-"`
	PID      string `json:"pid"`
	NameID   int64  `db:"node_name_id" json:"-"`
	Value    string `json:"value"`
	ValueURI string `db:"value_uri" json:"valueURI"`
}

// ListControlledValues gets all controlled values for a given name
func (db *DB) ListControlledValues(name string) []ControlledValue {
	var vals []ControlledValue
	err := db.Select(&vals,
		"SELECT cv.* FROM controlled_values cv inner join node_names nn on nn.id = cv.node_name_id WHERE nn.value=?", name)
	if err != nil {
		log.Printf("Unable to find controlled values for %s: %s", name, err.Error())
		return nil
	}
	return vals
}

// GetControlledValueByName finds a controlled value ecord by name
func (db *DB) GetControlledValueByName(name string) *ControlledValue {
	cv := ControlledValue{}
	err := db.Get(&cv, "SELECT * FROM controlled_values WHERE value=?", name)
	if err != nil {
		log.Printf("Unable to find controlled value %s: %s", name, err.Error())
		return nil
	}
	return &cv
}

// GetControlledValueByID finds a controlled value ecord by name
func (db *DB) GetControlledValueByID(id int64) *ControlledValue {
	cv := ControlledValue{}
	err := db.Get(&cv, "SELECT * FROM controlled_values WHERE id=?", id)
	if err != nil {
		log.Printf("Unable to find controlled value %d: %s", id, err.Error())
		return nil
	}
	return &cv
}
