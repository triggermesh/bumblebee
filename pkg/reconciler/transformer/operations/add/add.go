package add

import (
	"encoding/json"
	"strings"

	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/common/convert"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/common/storage"
)

// Add object implements Transformer interface.
type Add struct {
	Path  string
	Value string

	variables *storage.Storage
}

// runFirst is used to figure out if this operation should
// run before main Transformations. For example, Store
// operation needs to run first to load all Pipeline variables.
var runFirst bool = false

// operationName is used to identify this transformation.
var operationName string = "add"

// Register adds this transformation to the map which will
// be used to create Transformation pipeline.
func Register(m map[string]interface{}) {
	m[operationName] = &Add{}
}

// SetStorage sets a shared Storage with Pipeline variables.
func (a *Add) SetStorage(storage *storage.Storage) {
	a.variables = storage
}

// InitStep returns "true" if this Transformation should run
// as init step.
func (a *Add) InitStep() bool {
	return runFirst
}

// New returns a new instance of Add object.
func (a *Add) New(key, value string) interface{} {
	return &Add{
		Path:  key,
		Value: value,

		variables: a.variables,
	}
}

// Apply is a main method of Transformation that adds any type of
// variables into existing JSON.
func (a *Add) Apply(data []byte) ([]byte, error) {
	input := convert.SliceToMap(strings.Split(a.retrieveString(a.Path), "."), a.retrieveInterface(a.Value))
	event := make(map[string]interface{})
	if err := json.Unmarshal(data, &event); err != nil {
		return data, err
	}

	result := convert.MergeMaps(event, input)
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return data, err
	}

	return output, nil
}

func (a *Add) retrieveString(key string) string {
	if value, ok := a.variables.GetString(key); ok {
		return value
	}
	return key
}

func (a *Add) retrieveInterface(key string) interface{} {
	if value := a.variables.Get(key); value != nil {
		return value
	}
	return key
}
