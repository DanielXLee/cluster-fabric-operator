/*
© 2021 Red Hat, Inc. and others.

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

package utils

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/operator/common/embeddedyamls"
)

func CreateOrUpdate(c client.Client, obj client.Object) (bool, error) {
	// result, err := util.CreateOrUpdate(client, obj, util.Replace(obj))
	result, err := ctrl.CreateOrUpdate(context.TODO(), c, obj, func() error {
		return nil
	})
	return result == "created", err
}

func CreateOrUpdateClusterRole(c client.Client, clusterRole *rbacv1.ClusterRole) (bool, error) {
	return CreateOrUpdate(c, clusterRole)
}

func CreateOrUpdateClusterRoleBinding(c client.Client, clusterRoleBinding *rbacv1.ClusterRoleBinding) (bool, error) {
	return CreateOrUpdate(c, clusterRoleBinding)
}

func CreateOrUpdateCRD(c client.Client, crd *apiextensions.CustomResourceDefinition) (bool, error) {
	return CreateOrUpdate(c, crd)
}

func CreateOrUpdateEmbeddedCRD(c client.Client, crdYaml string) (bool, error) {
	crd := &apiextensions.CustomResourceDefinition{}

	if err := embeddedyamls.GetObject(crdYaml, crd); err != nil {
		return false, fmt.Errorf("error extracting embedded CRD: %s", err)
	}

	return CreateOrUpdateCRD(c, crd)
}

func CreateOrUpdateDeployment(c client.Client, deployment *appsv1.Deployment) (bool, error) {
	return CreateOrUpdate(c, deployment)
}

func CreateOrUpdateRole(c client.Client, role *rbacv1.Role) (bool, error) {
	return CreateOrUpdate(c, role)
}

func CreateOrUpdateRoleBinding(c client.Client, roleBinding *rbacv1.RoleBinding) (bool, error) {
	return CreateOrUpdate(c, roleBinding)
}

func CreateOrUpdateServiceAccount(c client.Client, sa *corev1.ServiceAccount) (bool, error) {
	return CreateOrUpdate(c, sa)
}
