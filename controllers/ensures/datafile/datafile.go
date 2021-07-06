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

package datafile

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/DanielXLee/cluster-fabric-operator/controllers/stringset"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/cluster"

	"github.com/DanielXLee/cluster-fabric-operator/controllers/components"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/broker"
)

type BrokerInfo struct {
	BrokerURL        string     `json:"brokerURL"`
	ClientToken      *v1.Secret `omitempty,json:"clientToken"`
	IPSecPSK         *v1.Secret `omitempty,json:"ipsecPSK"`
	ServiceDiscovery bool       `omitempty,json:"serviceDiscovery"`
	Components       []string   `json:",omitempty"`
	CustomDomains    *[]string  `omitempty,json:"customDomains"`
}

func (data *BrokerInfo) SetComponents(componentSet stringset.Interface) {
	data.Components = componentSet.Elements()
}

func (data *BrokerInfo) GetComponents() stringset.Interface {
	return stringset.New(data.Components...)
}

func (data *BrokerInfo) IsConnectivityEnabled() bool {
	return data.GetComponents().Contains(components.Connectivity)
}

func (data *BrokerInfo) IsServiceDiscoveryEnabled() bool {
	return data.GetComponents().Contains(components.ServiceDiscovery) || data.ServiceDiscovery
}

func (data *BrokerInfo) ToString() (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(jsonBytes), nil
}

func NewFromString(str string) (*BrokerInfo, error) {
	data := &BrokerInfo{}
	bytes, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return data, json.Unmarshal(bytes, data)
}

func (data *BrokerInfo) WriteToFile(filename string) error {
	dataStr, err := data.ToString()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filename, []byte(dataStr), 0600); err != nil {
		return err
	}

	return nil
}

func (data *BrokerInfo) WriteConfigMap(c client.Client, brokerNamespace string) error {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "broker-info",
			Namespace: brokerNamespace,
		},
	}
	or, err := ctrl.CreateOrUpdate(context.TODO(), c, cm, func() error {
		dataStr, err := data.ToString()
		if err != nil {
			return err
		}
		cm.Data = map[string]string{"brokerInfo": dataStr}
		return nil
	})
	if err != nil {
		return err
	}
	klog.Infof("Configmap broker-info %s", or)
	return nil
}

func NewFromFile(filename string) (*BrokerInfo, error) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewFromString(string(dat))
}

func NewFromConfigMap(c client.Client, brokerNamespace string) (*BrokerInfo, error) {
	cm := &v1.ConfigMap{}
	cmKey := types.NamespacedName{Name: "broker-info", Namespace: brokerNamespace}
	if err := c.Get(context.TODO(), cmKey, cm); err != nil {
		return nil, err
	}
	return NewFromString(cm.Data["brokerInfo"])
}

func NewFromCluster(c client.Client, restConfig *rest.Config, brokerNamespace string) (*BrokerInfo, error) {
	brokerInfo, err := newFromCluster(c, brokerNamespace)
	if err != nil {
		return nil, err
	}
	brokerInfo.BrokerURL = restConfig.Host + restConfig.APIPath
	return brokerInfo, err
}

func newFromCluster(c client.Client, brokerNamespace string) (*BrokerInfo, error) {
	brokerInfo := &BrokerInfo{}
	var err error

	brokerInfo.ClientToken, err = broker.GetClientTokenSecret(c, brokerNamespace, broker.SubmarinerBrokerAdminSA)
	if err != nil {
		return nil, err
	}
	brokerInfo.IPSecPSK, err = newIPSECPSKSecret()
	return brokerInfo, err
}

func (data *BrokerInfo) GetBrokerAdministratorCluster() (cluster.Cluster, error) {
	config := data.GetBrokerAdministratorConfig()
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	return cluster.New(config, func(clusterOptions *cluster.Options) {
		clusterOptions.Scheme = scheme
	})
}

func (data *BrokerInfo) GetBrokerAdministratorConfig() *rest.Config {
	tlsClientConfig := rest.TLSClientConfig{}
	if len(data.ClientToken.Data["ca.crt"]) != 0 {
		tlsClientConfig.CAData = data.ClientToken.Data["ca.crt"]
	}
	bearerToken := data.ClientToken.Data["token"]
	restConfig := rest.Config{
		Host:            data.BrokerURL,
		TLSClientConfig: tlsClientConfig,
		BearerToken:     string(bearerToken),
	}
	return &restConfig
}
