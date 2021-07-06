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

package broker

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/DanielXLee/cluster-fabric-operator/controllers/components"
	consts "github.com/DanielXLee/cluster-fabric-operator/controllers/ensures"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/gateway"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/lighthouse"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/utils"
	crdutils "github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/utils/crds"
)

func Ensure(c client.Client, config *rest.Config, componentArr []string, crds bool) error {
	if crds {
		crdCreator, err := crdutils.NewFromRestConfig(config)
		if err != nil {
			return fmt.Errorf("error accessing the target cluster: %s", err)
		}

		for i := range componentArr {
			switch componentArr[i] {
			case components.Connectivity:
				err = gateway.Ensure(c)
				if err != nil {
					return fmt.Errorf("error setting up the connectivity requirements: %s", err)
				}
			case components.ServiceDiscovery:
				_, err = lighthouse.Ensure(crdCreator, c, lighthouse.BrokerCluster)
				if err != nil {
					return fmt.Errorf("error setting up the service discovery requirements: %s", err)
				}
			case components.Globalnet:
				// Globalnet needs the Lighthouse CRDs too
				_, err = lighthouse.Ensure(crdCreator, c, lighthouse.BrokerCluster)
				if err != nil {
					return fmt.Errorf("error setting up the globalnet requirements: %s", err)
				}
			}
		}
	}

	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	return fmt.Errorf("error creating the core kubernetes clientset: %s", err)
	// }

	// Create the namespace
	err := CreateNewBrokerNamespace(c)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("error creating the broker namespace %s", err)
	}

	// Create administrator SA, Role, and bind them
	if err := createBrokerAdministratorRoleAndSA(c); err != nil {
		return err
	}

	// Create cluster Role, and a default account for backwards compatibility, also bind it
	if err := createBrokerClusterRoleAndDefaultSA(c); err != nil {
		return err
	}
	_, err = WaitForClientToken(c, SubmarinerBrokerAdminSA)
	return err
}

func createBrokerClusterRoleAndDefaultSA(c client.Client) error {
	// Create the a default SA for cluster access (backwards compatibility with documentation)
	err := CreateNewBrokerSA(c, submarinerBrokerClusterDefaultSA)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("error creating the default broker service account: %s", err)
	}

	// Create the broker cluster role, which will also be used by any new enrolled cluster
	_, err = CreateOrUpdateClusterBrokerRole(c)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("error creating broker role: %s", err)
	}

	// Create the role binding
	err = CreateNewBrokerRoleBinding(c, submarinerBrokerClusterDefaultSA, submarinerBrokerClusterRole)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("error creating the broker rolebinding: %s", err)
	}
	return nil
}

// CreateSAForCluster creates a new SA, and binds it to the submariner cluster role
func CreateSAForCluster(c client.Client, reader client.Reader, clusterID string) (*v1.Secret, error) {
	saName := fmt.Sprintf(submarinerBrokerClusterSAFmt, clusterID)
	err := CreateNewBrokerSA(c, saName)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("error creating cluster sa: %s", err)
	}

	err = CreateNewBrokerRoleBinding(c, saName, submarinerBrokerClusterRole)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("error binding sa to cluster role: %s", err)
	}

	clientToken, err := WaitForClientToken(reader, saName)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("error getting cluster sa token: %s", err)
	}
	return clientToken, nil
}

func createBrokerAdministratorRoleAndSA(c client.Client) error {
	// Create the SA we need for the managing the broker
	err := CreateNewBrokerSA(c, SubmarinerBrokerAdminSA)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		klog.Errorf("error creating the broker admin service account: %v", err)
		return err
	}

	// Create the broker admin role
	_, err = CreateOrUpdateBrokerAdminRole(c)
	if err != nil {
		klog.Errorf("error creating broker role: %v", err)
		return err
	}

	// Create the role binding
	err = CreateNewBrokerRoleBinding(c, SubmarinerBrokerAdminSA, submarinerBrokerAdminRole)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		klog.Errorf("error creating the broker rolebinding: %v", err)
		return err
	}
	return nil
}

func WaitForClientToken(reader client.Reader, submarinerBrokerSA string) (secret *v1.Secret, err error) {
	// wait for the client token to be ready, while implementing
	// exponential backoff pattern, it will wait a total of:
	// sum(n=0..9, 1.2^n * 5) seconds, = 130 seconds

	backoff := wait.Backoff{
		Steps:    10,
		Duration: 5 * time.Second,
		Factor:   1.2,
		Jitter:   1,
	}

	var lastErr error
	err = wait.ExponentialBackoff(backoff, func() (bool, error) {
		secret, lastErr = GetClientTokenSecret(reader, consts.SubmarinerBrokerNamespace, submarinerBrokerSA)
		if lastErr != nil {
			return false, nil
		}
		return true, nil
	})
	if err == wait.ErrWaitTimeout {
		return nil, lastErr
	}

	return secret, err
}

func CreateNewBrokerNamespace(c client.Client) error {
	return c.Create(context.TODO(), NewBrokerNamespace())
}

func CreateOrUpdateClusterBrokerRole(c client.Client) (created bool, err error) {
	return utils.CreateOrUpdateRole(c, NewBrokerClusterRole())
}

func CreateOrUpdateBrokerAdminRole(c client.Client) (created bool, err error) {
	return utils.CreateOrUpdateRole(c, NewBrokerAdminRole())
}

func CreateNewBrokerRoleBinding(c client.Client, serviceAccount, role string) error {
	return c.Create(context.TODO(), NewBrokerRoleBinding(serviceAccount, role))
}

func CreateNewBrokerSA(c client.Client, submarinerBrokerSA string) error {
	return c.Create(context.TODO(), NewBrokerSA(submarinerBrokerSA))
}
