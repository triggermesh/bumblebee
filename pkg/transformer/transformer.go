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
	"github.com/triggermesh/transformation/pkg/apis/transformation/v1alpha1"
	"github.com/triggermesh/transformation/pkg/transformer/common/storage"
	"github.com/triggermesh/transformation/pkg/transformer/operations"
)

// Transformer contains Pipelines for CE Context and Data transformations.
type Transformer struct {
	ContextPipeline *operations.Pipeline
	DataPipeline    *operations.Pipeline
}

// ceContext represents CloudEvents context structure but with exported Extensions.
type ceContext struct {
	*cloudevents.EventContextV1 `json:",inline"`
	Extensions                  map[string]interface{} `json:"Extensions,omitempty"`
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

	localContext := ceContext{
		EventContextV1: event.Context.AsV1(),
		Extensions:     event.Context.AsV1().GetExtensions(),
	}

	localContextBytes, err := json.Marshal(localContext)
	if err != nil {
		return &event, fmt.Errorf("cannot encode CE context: %w", err)
	}

	// Run init step such as load Pipeline variables first
	t.ContextPipeline.InitStep(localContextBytes)
	t.DataPipeline.InitStep(event.Data())

	// CE Context transformation
	localContextBytes, err = t.ContextPipeline.Apply(localContextBytes)
	if err != nil {
		return &event, fmt.Errorf("cannot apply transformation on CE context: %w", err)
	}

	if err := json.Unmarshal(localContextBytes, &localContext); err != nil {
		return &event, fmt.Errorf("cannot decode CE new context: %w", err)
	}
	event.Context = localContext
	for k, v := range localContext.Extensions {
		if err := event.Context.SetExtension(k, v); err != nil {
			log.Printf("Cannot set CE extension: %v", err)
		}
	}

	// CE Data transformation
	data, err := t.DataPipeline.Apply(event.Data())
	if err != nil {
		return &event, fmt.Errorf("cannot apply transformation on CE data: %w", err)
	}
	if err = event.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return &event, fmt.Errorf("cannot set data: %w", err)
	}

	return &event, nil
}
