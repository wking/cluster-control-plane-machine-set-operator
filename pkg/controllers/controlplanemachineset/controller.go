/*
Copyright 2022 Red Hat, Inc.

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

package controlplanemachineset

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	configv1 "github.com/openshift/api/config/v1"
	machinev1 "github.com/openshift/api/machine/v1"
	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	errorutils "k8s.io/apimachinery/pkg/util/errors"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// clusterControlPlaneMachineSetName is the name of the ControlPlaneMachineSet.
	// As ControlPlaneMachineSets are singletons within the namespace, only ControlPlaneMachineSets
	// with this name should be reconciled.
	clusterControlPlaneMachineSetName = "cluster"
)

// ControlPlaneMachineSetReconciler reconciles a ControlPlaneMachineSet object.
type ControlPlaneMachineSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// Namespace is the namespace in which the ControlPlaneMachineSet controller should operate.
	// Any ControlPlaneMachineSet not in this namespace should be ignored.
	Namespace string

	// OperatorName is the name of the ClusterOperator with which the controller should report
	// its status.
	OperatorName string
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlPlaneMachineSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&machinev1.ControlPlaneMachineSet{}, builder.WithPredicates(filterControlPlaneMachineSet(r.Namespace))).
		Owns(&machinev1beta1.Machine{}, builder.WithPredicates(filterControlPlaneMachines(r.Namespace))).
		Watches(&source.Kind{Type: &configv1.ClusterOperator{}}, handler.EnqueueRequestsFromMapFunc(clusterOperatorToControlPlaneMachineSet(r.Namespace, r.OperatorName))).
		Complete(r); err != nil {
		return fmt.Errorf("could not set up controller for ControlPlaneMachineSet: %w", err)
	}

	return nil
}

// Reconcile reconciles the ControlPlaneMachineSet object.
func (r *ControlPlaneMachineSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx, "namepace", req.Namespace, "name", req.Name)

	logger.V(1).Info("Reconciling control plane machine set")
	defer logger.V(1).Info("Finished reconciling control plane machine set")

	cpms := &machinev1.ControlPlaneMachineSet{}
	cpmsKey := client.ObjectKey{Namespace: req.Namespace, Name: req.Name}

	// Fetch the ControlPlaneMachineSet and set the cluster operator to available if it doesn't exist.
	if err := r.Get(ctx, cpmsKey, cpms); apierrors.IsNotFound(err) {
		logger.V(1).Info("No control plane machine set found, setting operator status available")

		if err := r.setClusterOperatorAvailable(ctx, logger); err != nil {
			return ctrl.Result{}, fmt.Errorf("unable to reconcile cluster operator status: %w", err)
		}
	} else if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to fetch control plane machine set: %w", err)
	}

	// Take a copy of the original object to be able to create a patch for the status at the end.
	patchBase := client.MergeFrom(cpms)

	// Collect errors as an aggregate to return together after all patches have been performed.
	var errs []error

	result, err := r.reconcile(ctx, logger, cpms)
	if err != nil {
		// Don't return an error here so that we have an opportunity to update the status and cluster operator status.
		errs = append(errs, fmt.Errorf("error reconciling control plane machine set: %w", err))
	}

	if err := r.updateControlPlaneMachineSetStatus(ctx, logger, cpms, patchBase); err != nil {
		// Don't return an error here so that we have an opportunity to update the cluster operator status.
		errs = append(errs, fmt.Errorf("error updating control plane machine set status: %w", err))
	}

	if err := r.updateClusterOperatorStatus(ctx, logger, cpms); err != nil {
		// Don't return an error here so we can aggregate the errors with previous updates.
		errs = append(errs, fmt.Errorf("error updating control plane machine set status: %w", err))
	}

	if len(errs) > 0 {
		return ctrl.Result{}, errorutils.NewAggregate(errs)
	}

	return result, nil
}

// reconcile performs the main business logic of the ControlPlaneMachineSet operator.
// Notably it actions the various parts of the business logic without performing any status updates on the
// ControlPlaneMachineSet object itself, these updates are handled at the parent scope.
func (r *ControlPlaneMachineSetReconciler) reconcile(ctx context.Context, logger logr.Logger, cpms *machinev1.ControlPlaneMachineSet) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}