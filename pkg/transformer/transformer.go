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

package transformer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/triggermesh/transformation-prototype/pkg/apis/transformation/v1alpha1"
	"github.com/triggermesh/transformation-prototype/pkg/transformer/common/storage"
	"github.com/triggermesh/transformation-prototype/pkg/transformer/operations"
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
	contextPipeline.SetStorage(sharedVars)
	dataPipeline.SetStorage(sharedVars)

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

	contextCE, err := json.Marshal(event.Context)
	if err != nil {
		return &event, fmt.Errorf("Cannot encode CE context: %v", err)
	}

	// Run init step such as load Pipeline variables first
	t.ContextPipeline.InitStep(contextCE)
	t.DataPipeline.InitStep(event.Data())

	// CE Context transformation
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