/*
Copyright 2021.

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

package controllers

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/datafile"

	operatorv1alpha1 "github.com/DanielXLee/cluster-fabric-operator/api/v1alpha1"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/components"
)

// FabricReconciler reconciles a Fabric object
type FabricReconciler struct {
	client.Client
	client.Reader
	*rest.Config
	Scheme       *runtime.Scheme
	DeployBroker bool
	JoinBroker   bool
}

const brokerDetailsFilename = "broker-info.subm"
const (
	SubmarinerBrokerNamespace = "submariner-k8s-broker"
)

// var defaultComponents = []string{components.ServiceDiscovery, components.Connectivity}
var validComponents = []string{components.ServiceDiscovery, components.Connectivity}

//+kubebuilder:rbac:groups=operator.tkestack.io,resources=fabrics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.tkestack.io,resources=fabrics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.tkestack.io,resources=fabrics/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Fabric object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *FabricReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.Infof("Start reconciling Fabric: %s", req.NamespacedName)
	instance := &operatorv1alpha1.Fabric{}

	if err := r.Client.Get(context.TODO(), req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	originalInstance := instance.DeepCopy()
	// Always attempt to patch the status after each reconciliation.
	defer func() {
		if reflect.DeepEqual(originalInstance.Status, instance.Status) {
			return
		}
		if updateErr := r.Status().Update(ctx, instance, &client.UpdateOptions{}); updateErr != nil {
			klog.Errorf("Update status failed, err: %v", updateErr)
		}
	}()

	// Deploy submeriner broker
	if r.DeployBroker {
		klog.Info("Deploy submeriner broker")
		if err := r.DeploySubmerinerBroker(instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Join managed cluster to submeriner borker
	if r.JoinBroker {
		klog.Info("Join managed cluster to submeriner broker")
		brokerInfo, err := datafile.NewFromConfigMap(r.Client, SubmarinerBrokerNamespace)
		if err != nil {
			return ctrl.Result{}, err
		}
		if err := r.JoinSubmarinerCluster(instance, brokerInfo); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FabricReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Fabric{}).
		Complete(r)
}
