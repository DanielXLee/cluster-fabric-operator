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

	createdRole, err := ensureRoles(c, namespace)
	if err != nil {
		return false, err
	}

	createdRB, err := ensureRoleBindings(c, namespace)
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

	return createdSA || createdRole || createdRB || createdCR || createdCRB, nil
}

func ensureServiceAccounts(c client.Client, namespace string) (bool, error) {
	createdOperatorSA, err := serviceaccount.Ensure(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_operator_service_account_yaml)
	if err != nil {
		return false, err
	}

	createdSubmarinerSA, err := serviceaccount.Ensure(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_gateway_service_account_yaml)
	if err != nil {
		return false, err
	}

	createdRouteAgentSA, err := serviceaccount.Ensure(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_route_agent_service_account_yaml)
	if err != nil {
		return false, err
	}

	createdGlobalnetSA, err := serviceaccount.Ensure(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_globalnet_service_account_yaml)
	if err != nil {
		return false, err
	}

	createdNPSyncerSA, err := serviceaccount.Ensure(c, namespace,
		embeddedyamls.Manifests_config_rbac_networkplugin_syncer_service_account_yaml)
	return createdOperatorSA || createdSubmarinerSA || createdRouteAgentSA || createdGlobalnetSA || createdNPSyncerSA, err
}

func ensureClusterRoles(c client.Client) (bool, error) {
	createdOperatorCR, err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_config_rbac_submariner_operator_cluster_role_yaml)
	if err != nil {
		return false, err
	}

	createdSubmarinerCR, err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_config_rbac_submariner_gateway_cluster_role_yaml)
	if err != nil {
		return false, err
	}

	createdRouteAgentCR, err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_config_rbac_submariner_route_agent_cluster_role_yaml)
	if err != nil {
		return false, err
	}

	createdGlobalnetCR, err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_config_rbac_submariner_globalnet_cluster_role_yaml)
	if err != nil {
		return false, err
	}

	createdNPSyncerCR, err := serviceaccount.EnsureClusterRole(c,
		embeddedyamls.Manifests_config_rbac_networkplugin_syncer_cluster_role_yaml)
	return createdOperatorCR || createdSubmarinerCR || createdRouteAgentCR || createdGlobalnetCR || createdNPSyncerCR, err
}

func ensureClusterRoleBindings(c client.Client, namespace string) (bool, error) {
	createdOperatorCRB, err := serviceaccount.EnsureClusterRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_operator_cluster_role_binding_yaml)
	if err != nil {
		return false, err
	}

	createdSubmarinerCRB, err := serviceaccount.EnsureClusterRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_gateway_cluster_role_binding_yaml)
	if err != nil {
		return false, err
	}

	createdRouteAgentCRB, err := serviceaccount.EnsureClusterRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_route_agent_cluster_role_binding_yaml)
	if err != nil {
		return false, err
	}

	createdGlobalnetCRB, err := serviceaccount.EnsureClusterRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_globalnet_cluster_role_binding_yaml)
	if err != nil {
		return false, err
	}

	createdNPSyncerCRB, err := serviceaccount.EnsureClusterRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_networkplugin_syncer_cluster_role_binding_yaml)
	return createdOperatorCRB || createdSubmarinerCRB || createdRouteAgentCRB || createdGlobalnetCRB || createdNPSyncerCRB, err
}

func ensureRoles(c client.Client, namespace string) (bool, error) {
	createdOperatorRole, err := serviceaccount.EnsureRole(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_operator_role_yaml)
	if err != nil {
		return false, err
	}

	createdSubmarinerRole, err := serviceaccount.EnsureRole(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_gateway_role_yaml)
	if err != nil {
		return false, err
	}

	createdRouteAgentRole, err := serviceaccount.EnsureRole(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_route_agent_role_yaml)
	if err != nil {
		return false, err
	}

	createdGlobalnetRole, err := serviceaccount.EnsureRole(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_globalnet_role_yaml)
	if err != nil {
		return false, err
	}

	return createdOperatorRole || createdSubmarinerRole || createdRouteAgentRole || createdGlobalnetRole, err
}

func ensureRoleBindings(c client.Client, namespace string) (bool, error) {
	createdOperatorRB, err := serviceaccount.EnsureRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_operator_role_binding_yaml)
	if err != nil {
		return false, err
	}

	createdSubmarinerRB, err := serviceaccount.EnsureRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_gateway_role_binding_yaml)
	if err != nil {
		return false, err
	}

	createdRouteAgentRB, err := serviceaccount.EnsureRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_route_agent_role_binding_yaml)
	if err != nil {
		return false, err
	}

	createdGlobalnetRB, err := serviceaccount.EnsureRoleBinding(c, namespace,
		embeddedyamls.Manifests_config_rbac_submariner_globalnet_role_binding_yaml)
	if err != nil {
		return false, err
	}

	return createdOperatorRB || createdSubmarinerRB || createdRouteAgentRB || createdGlobalnetRB, err
}
