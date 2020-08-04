package add

import "fmt"

type Add struct {
	Path  string
	Value string
}

var OperationName string = "add"

func Register(m map[string]interface{}) map[string]interface{} {
	m[OperationName] = Add{}
	return m
}

func (a Add) New(key, value string) interface{} {
	return Add{
		Path:  key,
		Value: value,
	}
}

func (a Add) Apply(data []byte) []byte {
	fmt.Printf("operation: %s\npaths: %v\n", OperationName, a)
	return data
}
