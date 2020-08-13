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

package operations

import (
	"fmt"
	"log"

	"github.com/triggermesh/transformation-prototype/pkg/apis/transformation/v1alpha1"
	"github.com/triggermesh/transformation-prototype/pkg/transformer/common/storage"
	"github.com/triggermesh/transformation-prototype/pkg/transformer/operations/add"
	"github.com/triggermesh/transformation-prototype/pkg/transformer/operations/delete"
	"github.com/triggermesh/transformation-prototype/pkg/transformer/operations/shift"
	"github.com/triggermesh/transformation-prototype/pkg/transformer/operations/store"
)

// Transformer is an interface that contains common methods
// to work with JSON data.
type Transformer interface {
	New(string, string) interface{}
	Apply([]byte) ([]byte, error)
	SetStorage(*storage.Storage)
	InitStep() bool
}

// Pipeline is a set of Transformations that are
// sequentially applied to JSON data.
type Pipeline struct {
	Transformers []Transformer
}

// register loads available Transformation into a named map.
func register() map[string]Transformer {
	m := make(map[string]interface{})

	add.Register(m)
	delete.Register(m)
	shift.Register(m)
	store.Register(m)

	transformations := make(map[string]Transformer)
	for k, v := range m {
		transformer, ok := v.(Transformer)
		if !ok {
			log.Printf("Operation %q doesn't implement Transformation interface, skipping", k)
			continue
		}
		transformations[k] = transformer
	}
	return transformations
}

// New loads available Transformations and creates a Pipeline.
func New(transformations []v1alpha1.Transform) (*Pipeline, error) {
	availableTransformers := register()
	p := []Transformer{}

	for _, transformation := range transformations {
		operation, exist := availableTransformers[transformation.Name]
		if !exist {
			return nil, fmt.Errorf("transformation %q not found", transformation.Name)
		}
		for _, kv := range transformation.Paths {
			tr := operation.New(kv.Key, kv.Value)
			p = append(p, tr.(Transformer))
			log.Printf("%s: %s", transformation.Name, kv.Key)
		}
	}

	return &Pipeline{
		Transformers: p,
	}, nil
}

// SetStorage injects shared storage with Pipeline vars.
func (p *Pipeline) SetStorage(s *storage.Storage) {
	for _, v := range p.Transformers {
		v.SetStorage(s)
	}
}

// InitStep runs Transformations that are marked as InitStep.
func (p *Pipeline) InitStep(data []byte) {
	for _, v := range p.Transformers {
		if !v.InitStep() {
			continue
		}
		if _, err := v.Apply(data); err != nil {
			log.Printf("Failed to apply Init step: %v", err)
		}
	}
}

// Apply applies Pipeline transformations.
func (p *Pipeline) Apply(data []byte) ([]byte, error) {
	var err error
	for _, v := range p.Transformers {
		if v.InitStep() {
			continue
		}
		data, err = v.Apply(data)
		if err != nil {
			return data, err
		}
	}
	return data, nil
}
