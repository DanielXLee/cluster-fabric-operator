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
	"fmt"

	submarinerv1a1 "github.com/submariner-io/submariner-operator/apis/submariner/v1alpha1"
	"k8s.io/klog/v2"

	"github.com/DanielXLee/cluster-fabric-operator/controllers/discovery/globalnet"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/broker"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/datafile"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/operator/brokercr"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/stringset"

	operatorv1alpha1 "github.com/DanielXLee/cluster-fabric-operator/api/v1alpha1"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/components"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/operator/submarinerop"
)

func (r *FabricReconciler) DeploySubmerinerBroker(instance *operatorv1alpha1.Fabric) error {
	brokerConfig := &instance.Spec.BrokerConfig
	componentSet := stringset.New(brokerConfig.ComponentArr...)

	if err := isValidComponents(componentSet); err != nil {
		klog.Errorf("Invalid components parameter: %v", err)
	}

	if brokerConfig.ServiceDiscoveryEnabled {
		componentSet.Add(components.ServiceDiscovery)
	}

	if brokerConfig.GlobalnetEnable {
		componentSet.Add(components.Globalnet)
	}

	if valid, err := isValidGlobalnetConfig(instance); !valid {
		klog.Errorf("Invalid GlobalCIDR configuration: %v", err)
	}

	klog.Info("Setting up broker RBAC")
	if err := broker.Ensure(r.Client, r.Config, brokerConfig.ComponentArr, false); err != nil {
		klog.Errorf("Error setting up broker RBAC: %v", err)
	}
	klog.Info("Deploying the Submariner operator")
	if err := submarinerop.Ensure(r.Client, r.Config, true); err != nil {
		klog.Errorf("Error deploying the operator: %v", err)
	}
	klog.Info("Deploying the broker")
	if err := brokercr.Ensure(r.Client, populateBrokerSpec(instance)); err != nil {
		klog.Errorf("Broker deployment failed: %v", err)
	}

	klog.Infof("Creating %s file", brokerDetailsFilename)

	// // If deploy-broker is retried we will attempt to re-use the existing IPsec PSK secret
	// if brokerConfig.IpsecSubmFile == "" {
	// 	if _, err := datafile.NewFromFile(brokerDetailsFilename); err == nil {
	// 		brokerConfig.IpsecSubmFile = brokerDetailsFilename
	// 		klog.Infof("Reusing IPsec PSK from existing %s", brokerDetailsFilename)
	// 	} else {
	// 		klog.Infof("A new IPsec PSK will be generated for %s", brokerDetailsFilename)
	// 	}
	// }

	subctlData, err := datafile.NewFromCluster(r.Client, r.Config, broker.SubmarinerBrokerNamespace)
	if err != nil {
		klog.Errorf("Error retrieving preparing the subm data file: %v", err)
	}
	newFilename, err := datafile.BackupIfExists(brokerDetailsFilename)
	if err != nil {
		klog.Errorf("Error backing up the brokerfile: %v", err)
	}
	if newFilename != "" {
		klog.Infof("Backed up previous %s to %s", brokerDetailsFilename, newFilename)
	}

	subctlData.ServiceDiscovery = brokerConfig.ServiceDiscoveryEnabled
	subctlData.SetComponents(componentSet)

	if len(brokerConfig.DefaultCustomDomains) > 0 {
		subctlData.CustomDomains = &brokerConfig.DefaultCustomDomains
	}

	if brokerConfig.GlobalnetEnable {
		err = globalnet.ValidateExistingGlobalNetworks(r.Reader, broker.SubmarinerBrokerNamespace)
		klog.Errorf("Error validating existing globalCIDR configmap", err)
	}

	if err = broker.CreateGlobalnetConfigMap(r.Client, brokerConfig.GlobalnetEnable, brokerConfig.GlobalnetCIDRRange,
		brokerConfig.DefaultGlobalnetClusterSize, broker.SubmarinerBrokerNamespace); err != nil {
		klog.Errorf("Error creating globalCIDR configmap on Broker: %v", err)
	}
	// if err = subctlData.WriteToFile(brokerDetailsFilename); err != nil {
	// 	klog.Errorf("Error writing the broker information: %v", err)
	// }
	if err = subctlData.WriteConfigMap(r.Client, SubmarinerBrokerNamespace); err != nil {
		klog.Errorf("Error writing the broker information: %v", err)
	}
	return nil
}

func isValidComponents(componentSet stringset.Interface) error {
	validComponentSet := stringset.New(validComponents...)

	if componentSet.Size() < 1 {
		return fmt.Errorf("at least one component must be provided for deployment")
	}

	for _, component := range componentSet.Elements() {
		if !validComponentSet.Contains(component) {
			return fmt.Errorf("unknown component: %s", component)
		}
	}

	return nil
}

func isValidGlobalnetConfig(instance *operatorv1alpha1.Fabric) (bool, error) {
	brokerConfig := &instance.Spec.BrokerConfig
	var err error
	if !brokerConfig.GlobalnetEnable {
		return true, nil
	}
	defaultGlobalnetClusterSize, err := globalnet.GetValidClusterSize(brokerConfig.GlobalnetCIDRRange, brokerConfig.DefaultGlobalnetClusterSize)
	if err != nil || defaultGlobalnetClusterSize == 0 {
		return false, err
	}
	return true, err
}

func populateBrokerSpec(instance *operatorv1alpha1.Fabric) submarinerv1a1.BrokerSpec {
	brokerConfig := instance.Spec.BrokerConfig
	brokerSpec := submarinerv1a1.BrokerSpec{
		GlobalnetEnabled:            brokerConfig.GlobalnetEnable,
		GlobalnetCIDRRange:          brokerConfig.GlobalnetCIDRRange,
		DefaultGlobalnetClusterSize: brokerConfig.DefaultGlobalnetClusterSize,
		Components:                  brokerConfig.ComponentArr,
		DefaultCustomDomains:        brokerConfig.DefaultCustomDomains,
	}
	return brokerSpec
}
