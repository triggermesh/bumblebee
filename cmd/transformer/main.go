package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/triggermesh/transformation-prototype/pkg/apis/transformation/v1alpha1"
	"github.com/triggermesh/transformation-prototype/pkg/reconciler/transformer"
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

	pipeline, err := transformer.NewPipeline(transformations.Transformations)
	if err != nil {
		log.Fatalf("cannot create transformation pipeline: %v", err)
	}

	pipeline.Start(ctx, ceClient)
}
