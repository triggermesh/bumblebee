package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/qntfy/kazaam"

	"github.com/triggermesh/transformation-prototype/pkg/apis/transformation/v1alpha1"
)

const (
	envVarName = "transformSpec"
)

// Handler accepts cloudevents through input channel,
// applies transformation and returns result
type Handler struct {
	kazaam     *kazaam.Kazaam
	targetType string
	input      chan cloudevents.Event
	result     chan output
}

type output struct {
	event cloudevents.Event
	err   error
}

type pipeline struct {
	typeHandlers    map[string]*Handler
	defaultHandlers map[string]*Handler
}

func (p pipeline) eventReceiver(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	if event.DataContentType() != cloudevents.ApplicationJSON {
		// return event unchanged if data is not JSON
		return &event, nil
	}

	for _, defaultHandler := range p.defaultHandlers {
		defaultHandler.input <- event
		result := <-defaultHandler.result
		if result.err != nil {
			return nil, result.err
		}
		event = result.event
	}

	typedHandler, exist := p.typeHandlers[event.Type()]
	if exist {
		typedHandler.input <- event
		result := <-typedHandler.result
		if result.err != nil {
			return nil, result.err
		}
		event = result.event
	}

	return &event, nil
}

func main() {
	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatal("Failed to create client, ", err)
	}

	envvar, _ := os.LookupEnv(envVarName)
	if envvar == "" {
		log.Fatal(fmt.Errorf("transformation spec is empty"))
	}

	var transformations []v1alpha1.EventTransformation

	if err = json.Unmarshal([]byte(envvar), &transformations); err != nil {
		log.Fatal("cannot unmarshal env: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	th := make(map[string]*Handler)
	dh := make(map[string]*Handler)
	for i, eventType := range transformations {
		handler, err := NewHandler(eventType.Transform, eventType.TargetType)
		if err != nil {
			log.Fatalf("cannot create kazaam handler: %v", err)
		}
		defer handler.Stop()

		go handler.Run(ctx)

		if eventType.CEType == "" {
			dh[strconv.Itoa(i)] = handler
		} else {
			th[eventType.CEType] = handler
		}
	}

	p := pipeline{
		typeHandlers:    th,
		defaultHandlers: dh,
	}

	log.Fatal(c.StartReceiver(context.Background(), p.eventReceiver))
}

// NewHandler creates an instance of handler{}
func NewHandler(transformSpec, targetType string) (*Handler, error) {
	kaz, err := kazaam.NewKazaam(transformSpec)
	if err != nil {
		return nil, err
	}

	return &Handler{
		kazaam:     kaz,
		targetType: targetType,
		input:      make(chan cloudevents.Event),
		result:     make(chan output),
	}, nil
}

// Stop closes Handlers channels
func (h *Handler) Stop() {
	close(h.input)
	close(h.result)
}

// Run starts transformation worker
func (h *Handler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-h.input:
			data, err := h.kazaam.Transform(event.Data())
			if err != nil {
				h.result <- output{
					err: err,
				}
				continue
			}
			err = event.SetData(cloudevents.ApplicationJSON, data)
			if err != nil {
				h.result <- output{
					err: err,
				}
				continue
			}
			if h.targetType != "" {
				event.Extensions()
				event.SetExtension("transformtype", event.Type)
				event.SetType(h.targetType)
			}
			h.result <- output{
				event: event,
				err:   nil,
			}
		}
	}
}
