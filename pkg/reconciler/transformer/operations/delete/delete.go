package delete

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/common/storage"
)

// Delete object implements Transformer interface.
type Delete struct {
	Path  string
	Value string

	variables *storage.Storage
}

// runFirst is used to figure out if this operation should
// run before main Transformations. For example, Store
// operation needs to run first to load all Pipeline variables.
var runFirst bool = false

// operationName is used to identify this transformation.
var operationName string = "delete"

// Register adds this transformation to the map which will
// be used to create Transformation pipeline.
func Register(m map[string]interface{}) {
	m[operationName] = &Delete{}
}

// SetStorage sets a shared Storage with Pipeline variables.
func (d *Delete) SetStorage(storage *storage.Storage) {
	d.variables = storage
}

// InitStep returns "true" if this Transformation should run
// as init step.
func (d *Delete) InitStep() bool {
	return runFirst
}

// New returns a new instance of Delete object.
func (d *Delete) New(key, value string) interface{} {
	return &Delete{
		Path:  key,
		Value: value,

		variables: d.variables,
	}
}

// Apply is a main method of Transformation that removed any type of
// variables from existing JSON.
func (d *Delete) Apply(data []byte) ([]byte, error) {
	d.Path = d.retrieveString(d.Path)
	d.Value = d.retrieveString(d.Value)

	result, err := d.parse(data, "", "")
	if err != nil {
		return data, err
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return data, err
	}

	return output, nil
}

func (d *Delete) retrieveString(key string) string {
	if value, ok := d.variables.GetString(key); ok {
		return value
	}
	return key
}

func (d *Delete) parse(data interface{}, key, path string) (interface{}, error) {
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
		output = nil
	default:
		log.Printf("unhandled type %T\n", value)
	}

	return output, nil
}

func (d *Delete) filter(path string, value interface{}) bool {
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

func (d *Delete) filterPath(path string) bool {
	return "."+d.Path == path
}

func (d *Delete) filterValue(value interface{}) bool {
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

func (d *Delete) filterPathAndValue(path string, value interface{}) bool {
	return d.filterPath(path) && d.filterValue(value)
}
