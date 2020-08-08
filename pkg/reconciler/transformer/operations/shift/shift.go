package shift

import (
	"encoding/json"
	"strings"

	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/common/convert"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/common/storage"
)

// Shift object implements Transformer interface.
type Shift struct {
	Path    string
	NewPath string
	Value   string

	variables *storage.Storage
}

const delimeter string = ":"

// runFirst is used to figure out if this operation should
// run before main Transformations. For example, Store
// operation needs to run first to load all Pipeline variables.
var runFirst bool = false

// operationName is used to identify this transformation.
var operationName string = "shift"

// Register adds this transformation to the map which will
// be used to create Transformation pipeline.
func Register(m map[string]interface{}) {
	m[operationName] = &Shift{}
}

// SetStorage sets a shared Storage with Pipeline variables.
func (s *Shift) SetStorage(storage *storage.Storage) {
	s.variables = storage
}

// InitStep returns "true" if this Transformation should run
// as init step.
func (s *Shift) InitStep() bool {
	return runFirst
}

// New returns a new instance of Shift object.
func (s *Shift) New(key, value string) interface{} {
	// doubtful scheme, review needed
	keys := strings.Split(key, delimeter)
	if len(keys) != 2 {
		return nil
	}
	return &Shift{
		Path:    keys[0],
		NewPath: keys[1],
		Value:   value,

		variables: s.variables,
	}
}

// Apply is a main method of Transformation that moves existing
// values to a new locations.
func (s *Shift) Apply(data []byte) ([]byte, error) {
	oldPath := convert.SliceToMap(strings.Split(s.retrieveString(s.Path), "."), "")

	event := make(map[string]interface{})
	if err := json.Unmarshal(data, &event); err != nil {
		return data, err
	}

	newEvent, value := extractValue(event, oldPath)
	if s.Value != "" {
		if !equal(convert.TryStringToJSONType(s.retrieveInterface(s.Value)), value) {
			return data, nil
		}
	}

	newPath := convert.SliceToMap(strings.Split(s.retrieveString(s.NewPath), "."), value)

	result := convert.MergeMaps(newEvent, newPath)
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return data, err
	}

	return output, nil
}

func (s *Shift) retrieveInterface(key string) interface{} {
	if value := s.variables.Get(key); value != nil {
		return value
	}
	return key
}

func (s *Shift) retrieveString(key string) string {
	if value, ok := s.variables.GetString(key); ok {
		return value
	}
	return key
}

func extractValue(source, path map[string]interface{}) (map[string]interface{}, interface{}) {
	var resultPath interface{}
	for k, v := range path {
		switch value := v.(type) {
		case float64, bool, string:
			if value == "" {
				m, ok := source[k]
				if !ok {
					break
				}
				resultPath = m
				delete(source, k)
			}
		case []interface{}:
			sourceArr, ok := source[k].([]interface{})
			if !ok {
				break
			}

			index := len(value) - 1
			if index >= len(sourceArr) {
				break
			}

			m, ok := value[index].(map[string]interface{})
			if ok {
				sourceArr[index], resultPath = extractValue(sourceArr[index].(map[string]interface{}), m)
				source[k] = sourceArr
				break
			}

			resultPath = sourceArr[index]
			source[k] = sourceArr[:index]
			if len(sourceArr) > index {
				source[k] = append(sourceArr[:index], sourceArr[index+1:]...)
			}

		case map[string]interface{}:
			if _, ok := source[k]; !ok {
				break
			}

			sourceMap, ok := source[k].(map[string]interface{})
			if !ok {
				break
			}
			source[k], resultPath = extractValue(sourceMap, value)
		case nil:
			source[k] = nil
		}
	}
	return source, resultPath
}

func equal(a, b interface{}) bool {
	switch value := b.(type) {
	case string:
		v, ok := a.(string)
		if ok && v == value {
			return true
		}
	case bool:
		v, ok := a.(bool)
		if ok && v == value {
			return true
		}
	case float64:
		v, ok := a.(float64)
		if ok && v == value {
			return true
		}
	}
	return false
}
