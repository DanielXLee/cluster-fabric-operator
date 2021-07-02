/*
© 2019 Red Hat, Inc. and others.

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

package serviceaccount

import (
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/operator/common/embeddedyamls"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/utils"
)

// Ensure creates the given service account
func Ensure(c client.Client, namespace, yaml string) (bool, error) {
	sa := &v1.ServiceAccount{}
	err := embeddedyamls.GetObject(yaml, sa)
	if err != nil {
		return false, err
	}

	return utils.CreateOrUpdateServiceAccount(c, namespace, sa)
}

func EnsureRole(c client.Client, namespace, yaml string) (bool, error) {
	role := &rbacv1.Role{}
	err := embeddedyamls.GetObject(yaml, role)
	if err != nil {
		return false, err
	}

	return utils.CreateOrUpdateRole(c, namespace, role)
}

func EnsureRoleBinding(c client.Client, namespace, yaml string) (bool, error) {
	roleBinding := &rbacv1.RoleBinding{}
	err := embeddedyamls.GetObject(yaml, roleBinding)
	if err != nil {
		return false, err
	}

	return utils.CreateOrUpdateRoleBinding(c, namespace, roleBinding)
}

func EnsureClusterRole(c client.Client, yaml string) (bool, error) {
	clusterRole := &rbacv1.ClusterRole{}
	err := embeddedyamls.GetObject(yaml, clusterRole)
	if err != nil {
		return false, err
	}

	return utils.CreateOrUpdateClusterRole(c, clusterRole)
}

func EnsureClusterRoleBinding(c client.Client, namespace, yaml string) (bool, error) {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	err := embeddedyamls.GetObject(yaml, clusterRoleBinding)
	if err != nil {
		return false, err
	}

	clusterRoleBinding.Subjects[0].Namespace = namespace
	return utils.CreateOrUpdateClusterRoleBinding(c, clusterRoleBinding)
}
