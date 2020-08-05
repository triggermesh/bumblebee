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
	result, err := d.parse(data, "", "")
	if err != nil {
		log.Println(err)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Printf("cannot marshal result: %v", err)
	}

	return output
}

func (d Delete) parse(data interface{}, key, path string) (interface{}, error) {
	output := make(map[string]interface{})
	switch value := data.(type) {
	case []byte:
		var m interface{}
		if err := json.Unmarshal(value, &m); err != nil {
			return nil, fmt.Errorf("unmarshal err: %v", err)
		}
		o, err := d.parse(m, key, path)
		if err != nil {
			return nil, fmt.Errorf("recursive call in []bytes case: %v", err)
		}
		return o, nil
	case float64, bool, string:
		return value, nil
	case []interface{}:
		slice := []interface{}{}
		for i, v := range value {
			o, err := d.parse(v, key, fmt.Sprintf("%s[%d]", path, i))
			if err != nil {
				return nil, fmt.Errorf("recursive call in []interface case: %v", err)
			}
			slice = append(slice, o)
		}
		return slice, nil
	case map[string]interface{}:
		for k, v := range value {
			subPath := fmt.Sprintf("%s.%s", path, k)
			if d.filter(subPath, v) {
				continue
			}
			o, err := d.parse(v, k, subPath)
			if err != nil {
				return nil, fmt.Errorf("recursive call in map[]interface case: %v", err)
			}
			output[k] = o
		}
	case nil:
		output[key] = nil
	default:
		log.Printf("unhandled type %T\n", value)
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
