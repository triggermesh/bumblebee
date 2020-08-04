package transformer

import (
	"context"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/triggermesh/transformation-prototype/pkg/apis/transformation/v1alpha1"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations"
)

type Pipeline struct {
	Transformers []Transformer
}

type Transformer interface {
	New(string, string) interface{}
	Apply([]byte) []byte
}

type Transformation struct {
	Name   string
	Fields []kv
}

type kv struct {
	key   string
	value string
}

func NewPipeline(transformations []v1alpha1.EventTransformation) (*Pipeline, error) {
	availableOperations := operations.Register()
	pipe := []Transformer{}

	for _, transformation := range transformations {
		op, exist := availableOperations[transformation.Name]
		if !exist {
			return nil, fmt.Errorf("transformation %q not found", transformation.Name)
		}
		operation := op.(Transformer)

		for _, kv := range transformation.Paths {
			tr := operation.New(kv.Key, kv.Value)
			pipe = append(pipe, tr.(Transformer))
		}
	}

	return &Pipeline{
		Transformers: pipe,
	}, nil
}

func (p *Pipeline) Start(ctx context.Context, ceClient cloudevents.Client) error {
	return ceClient.StartReceiver(ctx, p.receiveAndTransform)
}

func (p *Pipeline) receiveAndTransform(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	for _, tr := range p.Transformers {
		data := tr.Apply(event.Data())
		err := event.SetData(cloudevents.ApplicationJSON, data)
		if err != nil {
			return nil, fmt.Errorf("cannot set data: %v", err)
		}
	}
	event.SetType("ce.after.transformation")
	return &event, nil
}
