package store

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations/store/storage"
)

type Store struct {
	Path  string
	Value string

	variables *storage.Storage
}

var OperationName string = "store"

func Register(m map[string]interface{}) {
	m[OperationName] = &Store{}
}

func (s *Store) InjectVars(storage *storage.Storage) {
	s.variables = storage
}

func (s *Store) New(key, value string) interface{} {
	return &Store{
		Path:  key,
		Value: value,

		variables: s.variables,
	}
}

func (s *Store) Apply(data []byte) ([]byte, error) {
	path := sliceToMap(strings.Split(s.Value, "."), "")

	event := make(map[string]interface{})
	if err := json.Unmarshal(data, &event); err != nil {
		return data, err
	}

	value := getValue(event, path)
	s.variables.Set(s.Path, value)

	return data, nil
}

func getValue(source, path map[string]interface{}) interface{} {
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
				resultPath = getValue(sourceArr[index].(map[string]interface{}), m)
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
			resultPath = getValue(sourceMap, value)
		}
	}
	return resultPath
}

func sliceToMap(path []string, value interface{}) map[string]interface{} {
	var array bool
	var index int
	i := strings.Index(path[0], "[")
	if i > -1 && len(path[0]) > i+1 {
		indexStr := path[0][i+1 : len(path[0])-1]
		indexInt, err := strconv.Atoi(indexStr)
		if err == nil {
			index = indexInt
			array = true
			path[0] = path[0][:i]
		}
	}

	if len(path) == 1 {
		if !array {
			return map[string]interface{}{
				path[0]: value,
			}
		}
		arr := make([]interface{}, index+1, index+1)
		arr[index] = value
		return map[string]interface{}{
			path[0]: arr,
		}
	}

	key := path[0]
	path = path[1:]
	m := sliceToMap(path, value)
	if !array {
		return map[string]interface{}{
			key: m,
		}
	}
	arr := make([]interface{}, index+1, index+1)
	arr[index] = m
	return map[string]interface{}{
		key: arr,
	}
}
