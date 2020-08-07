package operations

import (
	"log"

	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations/add"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations/delete"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations/shift"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations/store"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations/store/storage"
)

type Transformer interface {
	New(string, string) interface{}
	Apply([]byte) ([]byte, error)
	InjectVars(*storage.Storage)
}

func Register() map[string]Transformer {
	m := make(map[string]interface{})

	add.Register(m)
	delete.Register(m)
	shift.Register(m)
	store.Register(m)

	s := storage.New()
	transformations := make(map[string]Transformer)
	for k, v := range m {
		transformer, ok := v.(Transformer)
		if !ok {
			log.Printf("Operation %q doesn't implement Transformation interface, skipping", k)
			continue
		}
		transformer.InjectVars(s)
		transformations[k] = transformer
	}

	return transformations
}
