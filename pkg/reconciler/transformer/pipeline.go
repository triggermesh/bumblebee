package transformer

import (
	"context"
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/triggermesh/transformation-prototype/pkg/apis/transformation/v1alpha1"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations"
)

// Pipeline is a set of Transformations that that are
// sequentially applied to JSON data.
type Pipeline struct {
	Transformers []operations.Transformer
}

// NewPipeline loads available Transformations and creates a Pipeline.
func NewPipeline(transformations []v1alpha1.EventTransformation) (*Pipeline, error) {
	availableTransformers := operations.Register()
	pipe := []operations.Transformer{}

	for _, transformation := range transformations {
		operation, exist := availableTransformers[transformation.Name]
		if !exist {
			return nil, fmt.Errorf("transformation %q not found", transformation.Name)
		}
		for _, kv := range transformation.Paths {
			tr := operation.New(kv.Key, kv.Value)
			pipe = append(pipe, tr.(operations.Transformer))
			log.Printf("%s: %q loaded\n", transformation.Name, kv.Key)
		}
	}

	return &Pipeline{
		Transformers: pipe,
	}, nil
}

// Start runs CloudEvent receiver and applies Transformation pipeline
// on incoming events.
func (p *Pipeline) Start(ctx context.Context, ceClient cloudevents.Client) error {
	log.Println("Starting CloudEvent receiver")
	return ceClient.StartReceiver(ctx, p.receiveAndTransform)
}

func (p *Pipeline) receiveAndTransform(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	log.Printf("Received %q event\n", event.Type())
	for _, tr := range p.Transformers {
		data, err := tr.Apply(event.Data())
		if err != nil {
			log.Printf("Cannot apply transformation: %v", err)
		}
		if err = event.SetData(cloudevents.ApplicationJSON, data); err != nil {
			return nil, fmt.Errorf("cannot set data: %v", err)
		}
	}
	// TODO: handle CE attributes same way as a JSON payload.
	event.SetType("ce.after.transformation")
	return &event, nil
}
