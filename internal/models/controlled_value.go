package models

// ControlledValue is a controlled vocabulary for node values
type ControlledValue struct {
	ID    int64
	PID   string
	Name  NodeName
	Value string
}
