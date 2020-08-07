package store

import (
	"encoding/json"
	"strings"

	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/common/convert"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/common/storage"
)

// Store object implements Transformer interface.
type Store struct {
	Path  string
	Value string

	variables *storage.Storage
}

// OperationName is used to identify this transformation.
var OperationName string = "store"

// Register adds this transformation to the map which will
// be used to create Transformation pipeline.
func Register(m map[string]interface{}) {
	m[OperationName] = &Store{}
}

// InjectVars sets a shared Storage with Pipeline variables.
func (s *Store) InjectVars(storage *storage.Storage) {
	s.variables = storage
}

// New returns a new instance of Store object.
func (s *Store) New(key, value string) interface{} {
	return &Store{
		Path:  key,
		Value: value,

		variables: s.variables,
	}
}

// Apply is a main method of Transformation that stores JSON values
// into variables that can be used by other Transformations in a pipeline.
func (s *Store) Apply(data []byte) ([]byte, error) {
	path := convert.SliceToMap(strings.Split(s.retrieveString(s.Value), "."), "")

	event := make(map[string]interface{})
	if err := json.Unmarshal(data, &event); err != nil {
		return data, err
	}

	value := readValue(event, path)
	s.variables.Set(s.Path, value)

	return data, nil
}

func (s *Store) retrieveString(key string) string {
	if value, ok := s.variables.GetString(key); ok {
		return value
	}
	return key
}

func readValue(source, path map[string]interface{}) interface{} {
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
				resultPath = readValue(sourceArr[index].(map[string]interface{}), m)
				break
			}

			resultPath = sourceArr[index]

		case map[string]interface{}:
			if _, ok := source[k]; !ok {
				break
			}

			sourceMap, ok := source[k].(map[string]interface{})
			if !ok {
				break
			}
			resultPath = readValue(sourceMap, value)
		case nil:
			resultPath = nil
		}
	}
	return resultPath
}
