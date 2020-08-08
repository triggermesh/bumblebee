package transformer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/triggermesh/transformation-prototype/pkg/apis/transformation/v1alpha1"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/common/storage"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer/operations"
)

// Transformer contains Pipelines for CE Context and Data transformations.
type Transformer struct {
	ContextPipeline *operations.Pipeline
	DataPipeline    *operations.Pipeline
}

// NewTransformer creates Transformer instance.
func NewTransformer(context, data []v1alpha1.Transform) (Transformer, error) {
	contextPipeline, err := operations.New(context)
	if err != nil {
		return Transformer{}, err
	}

	dataPipeline, err := operations.New(data)
	if err != nil {
		return Transformer{}, err
	}

	sharedVars := storage.New()
	contextPipeline.InjectVars(sharedVars)
	dataPipeline.InjectVars(sharedVars)

	return Transformer{
		ContextPipeline: contextPipeline,
		DataPipeline:    dataPipeline,
	}, nil
}

// Start runs CloudEvent receiver and applies transformation Pipeline
// on incoming events.
func (t *Transformer) Start(ctx context.Context, ceClient cloudevents.Client) error {
	log.Println("Starting CloudEvent receiver")
	return ceClient.StartReceiver(ctx, t.receiveAndTransform)
}

func (t *Transformer) receiveAndTransform(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	if event.DataContentType() != cloudevents.ApplicationJSON {
		return &event, nil
	}
	log.Printf("Received %q event\n", event.Type())

	// CE Context transformation
	contextCE, err := json.Marshal(event.Context)
	if err != nil {
		return &event, fmt.Errorf("Cannot encode CE context: %v", err)

	}
	contextCE, err = t.ContextPipeline.Apply(contextCE)
	if err != nil {
		return &event, fmt.Errorf("Cannot apply transformation on CE context: %v", err)
	}
	newContext := cloudevents.EventContextV1{}
	if err := json.Unmarshal(contextCE, &newContext); err != nil {
		return &event, fmt.Errorf("Cannot decode CE new context: %v", err)
	}
	event.Context = newContext.AsV1()

	// CE Data transformation
	data, err := t.DataPipeline.Apply(event.Data())
	if err != nil {
		return &event, fmt.Errorf("Cannot apply transformation on CE data: %v", err)
	}
	if err = event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return &event, fmt.Errorf("cannot set data: %v", err)
	}

	return &event, nil
}
