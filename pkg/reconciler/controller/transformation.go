/*
Copyright 2020 Triggermesh Inc.

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

package controller

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/network"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"
	servingv1client "knative.dev/serving/pkg/client/clientset/versioned"
	servingv1listers "knative.dev/serving/pkg/client/listers/serving/v1"

	transformationv1alpha1 "github.com/triggermesh/bumblebee/pkg/apis/transformation/v1alpha1"
	transformationreconciler "github.com/triggermesh/bumblebee/pkg/client/generated/injection/reconciler/transformation/v1alpha1/transformation"
	"github.com/triggermesh/bumblebee/pkg/reconciler/controller/resources"
)

const (
	envVarName = "transformSpec"
)

// newReconciledNormal makes a new reconciler event with event type Normal, and
// reason AddressableServiceReconciled.
func newReconciledNormal(namespace, name string) reconciler.Event {
	return reconciler.NewEvent(corev1.EventTypeNormal, "TransformationReconciled", "Transformation reconciled: \"%s/%s\"", namespace, name)
}

// Reconciler implements addressableservicereconciler.Interface for
// Transformation resources.
type Reconciler struct {
	// Tracker builds an index of what resources are watching other resources
	// so that we can immediately react to changes tracked resources.
	Tracker tracker.Interface

	// Listers index properties about resources
	knServiceLister  servingv1listers.ServiceLister
	servingClientSet servingv1client.Interface

	transformerImage string
}

// Check that our Reconciler implements Interface
var _ transformationreconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, t *transformationv1alpha1.Transformation) reconciler.Event {
	logger := logging.FromContext(ctx)

	if err := r.Tracker.TrackReference(tracker.Reference{
		APIVersion: "serving.knative.dev/v1",
		Kind:       "Service",
		Name:       t.Name,
		Namespace:  t.Namespace,
	}, t); err != nil {
		logger.Errorf("Error tracking service %s: %v", t.Name, err)
		return err
	}

	// Reconcile Transformation and then write back any status updates regardless of
	// whether the reconcile error out.
	reconcileErr := r.reconcile(ctx, t)
	if reconcileErr != nil {
		logger.Error("Error reconciling Transformation", zap.Error(reconcileErr))
		return reconcileErr
	}

	logger.Debug("Transformation reconciled")
	return newReconciledNormal(t.Namespace, t.Name)
}

func (r *Reconciler) reconcile(ctx context.Context, t *transformationv1alpha1.Transformation) error {
	logger := logging.FromContext(ctx)

	trn, err := json.Marshal(t.Spec)
	if err != nil {
		logger.Errorf("Cannot encode transformation spec: %v", err)
		return nil
	}

	svc, err := r.knServiceLister.Services(t.Namespace).Get(t.Name)
	if apierrs.IsNotFound(err) {
		svc = resources.NewKnService(t.Namespace, t.Name,
			resources.Image(r.transformerImage),
			resources.EnvVar(envVarName, string(trn)),
			resources.KsvcLabelVisibilityClusterLocal(),
			resources.Owner(t),
		)
		_, err := r.servingClientSet.ServingV1().Services(t.Namespace).Create(svc)
		if err != nil {
			logger.Errorf("Cannot create kn service: %v", err)
			t.Status.MarkServiceUnavailable(t.Name)
			return err
		}
		logger.Info("Kn service created")
		return nil
	} else if err != nil {
		logger.Errorf("Error reconciling service %s: %v", t.Name, err)
		return err
	}

	if !resources.KnServiceHasEnvVar(svc, envVarName, string(trn)) ||
		!resources.KnServiceImage(svc, r.transformerImage) {
		logger.Info("Kn service spec outdated, updating service")
		newSvc := resources.NewKnService(t.Namespace, t.Name,
			resources.Image(r.transformerImage),
			resources.EnvVar(envVarName, string(trn)),
			resources.KsvcLabelVisibilityClusterLocal(),
		)
		svc.Spec = newSvc.Spec
		_, err := r.servingClientSet.ServingV1().Services(t.Namespace).Update(svc)
		if err != nil {
			logger.Errorf("Cannot update kn service: %v", err)
			t.Status.MarkServiceUnavailable(t.Name)
			return err
		}
	}

	if svc.IsReady() {
		t.Status.Address = &duckv1.Addressable{
			URL: &apis.URL{
				Scheme: "http",
				Host:   network.GetServiceHostname(t.Name, t.Namespace),
			},
		}
		t.Status.MarkServiceAvailable()
	}
	t.Status.CloudEventAttributes = r.createCloudEventAttributes(&t.Spec)

	return nil
}

func (r *Reconciler) createCloudEventAttributes(ts *transformationv1alpha1.TransformationSpec) []duckv1.CloudEventAttributes {
	ceAttributes := make([]duckv1.CloudEventAttributes, 0)
	for _, item := range ts.Context {
		if item.Operation == "add" {
			attribute := duckv1.CloudEventAttributes{}
			for _, path := range item.Paths {
				switch path.Key {
				case "type":
					attribute.Type = path.Value
				case "source":
					attribute.Source = path.Value
				}
			}
			if attribute.Source != "" || attribute.Type != "" {
				ceAttributes = append(ceAttributes, attribute)
			}
			break
		}
	}
	return ceAttributes
}
