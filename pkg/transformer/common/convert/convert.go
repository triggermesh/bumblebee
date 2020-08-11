/*
Copyright (c) 2020 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package convert

import (
	"strconv"
	"strings"
)

// TryStringToJSONType accepts interface value and if value is string
// it will try to format it into JSON friendly representation of bool
// or float64. Otherwise value will be returned unchanged.
func TryStringToJSONType(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f
		}
	}
	return value
}

// SliceToMap converts string slice into map that can be encoded into JSON.
func SliceToMap(path []string, value interface{}) map[string]interface{} {
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
				path[0]: TryStringToJSONType(value),
			}
		}
		arr := make([]interface{}, index+1)
		arr[index] = TryStringToJSONType(value)
		return map[string]interface{}{
			path[0]: arr,
		}
	}

	key := path[0]
	path = path[1:]
	m := SliceToMap(path, value)
	if !array {
		return map[string]interface{}{
			key: m,
		}
	}
	arr := make([]interface{}, index+1)
	arr[index] = m
	return map[string]interface{}{
		key: arr,
	}
}

// MergeMaps accepts two maps (effectively, JSONs) and merges them together.
// Source map keys are being overwritten by appendix keys if they overlap.
func MergeMaps(source, appendix map[string]interface{}) map[string]interface{} {
	for k, v := range appendix {
		switch value := v.(type) {
		case float64, bool, string, nil:
			if source == nil {
				source = make(map[string]interface{})
			}
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
			resArr := make([]interface{}, resArrLen)
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
			source[k] = MergeMaps(m, value)
		}
	}
	return source
}
