package shift

import (
	"encoding/json"
	"strconv"
	"strings"
)

type Shift struct {
	Path    string
	NewPath string
	Value   string
}

const delimeter string = ":"

var OperationName string = "shift"

func Register(m map[string]interface{}) map[string]interface{} {
	m[OperationName] = Shift{}
	return m
}

func (s Shift) New(key, value string) interface{} {
	// doubtful scheme, review needed
	keys := strings.Split(key, delimeter)
	if len(keys) != 2 {
		return nil
	}
	return Shift{
		Path:    keys[0],
		NewPath: keys[1],
		Value:   value,
	}
}

func (s Shift) Apply(data []byte) ([]byte, error) {
	oldPath := sliceToMap(strings.Split(s.Path, "."), "")

	event := make(map[string]interface{})
	if err := json.Unmarshal(data, &event); err != nil {
		return data, err
	}

	newEvent, value := getValue(event, oldPath)
	if s.Value != "" {
		if !equal(tryConvert(s.Value), value) {
			return data, nil
		}
	}

	newPath := sliceToMap(strings.Split(s.NewPath, "."), value)

	result := mergeMaps(newEvent, newPath)
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return data, err
	}

	return output, nil
}

func getValue(source, path map[string]interface{}) (map[string]interface{}, interface{}) {
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
				sourceArr[index], resultPath = getValue(sourceArr[index].(map[string]interface{}), m)
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
			source[k], resultPath = getValue(sourceMap, value)
		}
	}
	return source, resultPath
}

func mergeMaps(source, appendix map[string]interface{}) map[string]interface{} {
	for k, v := range appendix {
		switch value := v.(type) {
		case float64, bool, string:
			source[k] = value
		case []interface{}:
			sourceArr, ok := source[k].([]interface{})
			if !ok {
				source[k] = value
				return source
			}
			resArrLen := len(sourceArr)
			if len(value) > resArrLen {
				resArrLen = len(value)
			}
			resArr := make([]interface{}, resArrLen, resArrLen)
			for i := range resArr {
				if i < len(value) && value[i] != nil {
					resArr[i] = value[i]
					continue
				}
				if i < len(sourceArr) {
					resArr[i] = sourceArr[i]
				}
			}
			source[k] = resArr
		case map[string]interface{}:
			m, ok := source[k].(map[string]interface{})
			if !ok {
				m = make(map[string]interface{})
			}
			source[k] = mergeMaps(m, value)
		}
	}
	return source
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

func tryConvert(value string) interface{} {
	b, err := strconv.ParseBool(value)
	if err == nil {
		return b
	}
	f, err := strconv.ParseFloat(value, 64)
	if err == nil {
		return f
	}
	return value
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
