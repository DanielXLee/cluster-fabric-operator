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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/operator/common/serviceaccount"

	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/operator/common/embeddedyamls"
)

// Ensure functions updates or installs the operator CRDs in the cluster
func Ensure(c client.Client, namespace string) (bool, error) {
	createdSA, err := ensureServiceAccounts(c, namespace)
	if err != nil {
		return false, err
	}

	createdCR, err := ensureClusterRoles(c)
	if err != nil {
		return false, err
	}

	createdCRB, err := ensureClusterRoleBindings(c, namespace)
	if err != nil {
		return false, err
	}

	return createdSA || createdCR || createdCRB, nil
}

func ensureServiceAccounts(c client.Client, namespace string) (bool, error) {
	createdAgentSA, err := serviceaccount.Ensure(c, namespace,
		embeddedyamls.Manifests_config_rbac_lighthouse_agent_service_account_yaml)
	if err != nil {
		return false, err
	}

	createdCoreDNSSA, err := serviceaccount.Ensure(c, namespace,
		embeddedyamls.Manifests_config_rbac_lighthouse_coredns_service_account_yaml)
	if err != nil {
		return false, err
	}
	return createdAgentSA || createdCoreDNSSA, err
}

func ensureClusterRoles(c client.Client) (bool, error) {
	createdAgentCR, err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_config_rbac_lighthouse_agent_cluster_role_yaml)
	if err != nil {
		return false, err
	}

	createdCoreDNSCR, err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_config_rbac_lighthouse_coredns_cluster_role_yaml)
	if err != nil {
		return false, err
	}

	return createdAgentCR || createdCoreDNSCR, err
}

func ensureClusterRoleBindings(c client.Client, namespace string) (bool, error) {
	createdAgentCRB, err := serviceaccount.EnsureClusterRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_lighthouse_agent_cluster_role_binding_yaml)
	if err != nil {
		return false, err
	}

	createdCoreDNSCRB, err := serviceaccount.EnsureClusterRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_lighthouse_coredns_cluster_role_binding_yaml)
	if err != nil {
		return false, err
	}

	return createdAgentCRB || createdCoreDNSCRB, err
}