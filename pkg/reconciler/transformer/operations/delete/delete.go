package delete

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

type Delete struct {
	Path  string
	Value string
}

var OperationName string = "delete"

func Register(m map[string]interface{}) map[string]interface{} {
	m[OperationName] = Delete{}
	return m
}

func (d Delete) New(key, value string) interface{} {
	return Delete{
		Path:  key,
		Value: value,
	}
}

func (d Delete) Apply(data []byte) []byte {
	// fmt.Printf("operation: %s\npaths: %v\n", OperationName, d)
	result, err := d.parse(data, "")
	if err != nil {
		log.Println(err)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Printf("cannot marshal result: %v", err)
	}

	// fmt.Println(string(output))
	return output
}

func (d Delete) parse(data []byte, curPath string) (interface{}, error) {
	var m map[string]interface{}

	if err := json.Unmarshal(data, &m); err != nil {
		return string(data), nil
	}

	output := make(map[string]interface{})
	for k, v := range m {
		path := fmt.Sprintf("%s.%s", curPath, k)
		if d.filter(path, v) {
			// fmt.Println("filtered: ", path)
			continue
		}
		// fmt.Println(path)
		switch value := v.(type) {
		case float64, bool, string:
			output[k] = value
		case map[string]interface{}:
			data, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("marshal map[string]interface{}: %v", err)
			}
			o, err := d.parse(data, path)
			if err != nil {
				return nil, fmt.Errorf("recursive call: %v", err)
			}
			output[k] = o
		case []interface{}:
			slice := []interface{}{}
			for i, v := range value {
				data, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("marshal []interface{}: %v", err)
				}
				o, err := d.parse(data, fmt.Sprintf("%s[%d]", path, i))
				if err != nil {
					return nil, fmt.Errorf("recursive call: %v", err)
				}
				slice = append(slice, o)
			}
			output[k] = slice
		case nil:
			output[k] = nil
		default:
			log.Printf("unhandled type %T\n", value)
		}
	}

	return output, nil
}

func (d Delete) filter(path string, value interface{}) bool {
	switch {
	case d.Path != "" && d.Value != "":
		return d.filterPathAndValue(path, value)
	case d.Path != "":
		return d.filterPath(path)
	case d.Value != "":
		return d.filterValue(value)
	}
	return false
}

func (d Delete) filterPath(path string) bool {
	return "."+d.Path == path
}

func (d Delete) filterValue(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return v == d.Value
	case float64:
		return d.Value == strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return d.Value == fmt.Sprintf("%t", v)
	}
	return false
}

func (d Delete) filterPathAndValue(path string, value interface{}) bool {
	return d.filterPath(path) && d.filterValue(value)
}
