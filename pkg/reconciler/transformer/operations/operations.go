package operations

import (
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations/add"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations/delete"
)

func Register() map[string]interface{} {
	m := make(map[string]interface{})

	add.Register(m)
	delete.Register(m)

	return m
}
