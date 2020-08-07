package operations

import (
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations/add"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations/delete"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations/shift"
)

func Register() map[string]interface{} {
	m := make(map[string]interface{})

	add.Register(m)
	delete.Register(m)
	shift.Register(m)

	return m
}
