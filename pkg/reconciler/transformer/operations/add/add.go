package add

import (
	"encoding/json"
	"strconv"
	"strings"
)

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

func (a Add) Apply(data []byte) ([]byte, error) {
	input := sliceToMap(strings.Split(a.Path, "."), a.Value)
	event := make(map[string]interface{})
	if err := json.Unmarshal(data, &event); err != nil {
		return data, err
	}

	result := mergeMaps(event, input)
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return data, err
	}

	return output, nil
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

func sliceToMap(path []string, value string) map[string]interface{} {
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
				path[0]: tryConvert(value),
			}
		}
		arr := make([]interface{}, index+1, index+1)
		arr[index] = tryConvert(value)
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
