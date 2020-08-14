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

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/triggermesh/transformation-prototype/pkg/apis/transformation/v1alpha1"
	"github.com/triggermesh/transformation-prototype/pkg/transformer"
)

const (
	envVarName = "transformSpec"
)

func main() {
	ceClient, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}

	envvar, _ := os.LookupEnv(envVarName)
	if envvar == "" {
		log.Fatal("transformation spec is empty")
	}

	var transformations v1alpha1.TransformationSpec

	if err = json.Unmarshal([]byte(envvar), &transformations); err != nil {
		log.Fatalf("cannot unmarshal env: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	transformer, err := transformer.NewTransformer(transformations.Context, transformations.Data)
	if err != nil {
		log.Fatalf("cannot create transformation pipeline: %v", err)
	}

	log.Fatalf("Cannot start transformation listener: %v", transformer.Start(ctx, ceClient))
}
