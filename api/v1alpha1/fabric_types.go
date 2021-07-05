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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FabricSpec defines the desired state of Fabric
type FabricSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// BrokerConfig represents the broker cluster configuration of the Submariner.
	// +optional
	BrokerConfig `json:"brokerConfig,omitempty"`

	// JoinConfig represents the managed cluster join configuration of the Submariner.
	// +optional
	JoinConfig `json:"joinConfig,omitempty"`

	CloudPrepareConfig `json:"cloudPrepareConfig,omitempty"`
}

// FabricStatus defines the observed state of Fabric
type FabricStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

type BrokerConfig struct {
	IpsecSubmFile               string        `json:"ipsecSubmFile,omitempty"`
	GlobalnetEnable             bool          `json:"globalnetEnable,omitempty"`
	GlobalnetCIDRRange          string        `json:"globalnetCIDRRange,omitempty"`
	DefaultGlobalnetClusterSize uint          `json:"defaultGlobalnetClusterSize,omitempty"`
	ServiceDiscoveryEnabled     bool          `json:"serviceDiscoveryEnabled,omitempty"`
	ComponentArr                []string      `json:"componentArr,omitempty"`
	GlobalCIDRConfigMap         *v1.ConfigMap `json:"globalCIDRConfigMap,omitempty"`
	DefaultCustomDomains        []string      `json:"defaultCustomDomains,omitempty"`
}

type JoinConfig struct {
	ClusterID     string `json:"clusterID,omitempty"`
	ServiceCIDR   string `json:"serviceCIDR,omitempty"`
	ClusterCIDR   string `json:"clusterCIDR,omitempty"`
	GlobalnetCIDR string `json:"globalnetCIDR,omitempty"`
	Repository    string `json:"repository,omitempty"`
	ImageVersion  string `json:"imageVersion,omitempty"`
	// NattPort represents IPsec NAT-T port (default 4500).
	// +optional
	// +kubebuilder:default=4500

	NattPort int `json:"nattPort,omitempty"`
	// IkePort represents IPsec IKE port (default 500).
	// +optional
	// +kubebuilder:default=500
	IkePort                       int        `json:"ikePort,omitempty"`
	PreferredServer               bool       `json:"preferredServer,omitempty"`
	ForceUDPEncaps                bool       `json:"forceUDPEncaps,omitempty"`
	NatTraversal                  bool       `json:"natTraversal,omitempty"`
	GlobalnetEnabled              bool       `json:"globalnetEnabled,omitempty"`
	IpsecDebug                    bool       `json:"ipsecDebug,omitempty"`
	SubmarinerDebug               bool       `json:"submarinerDebug,omitempty"`
	LabelGateway                  bool       `json:"labelGateway,omitempty"`
	LoadBalancerEnabled           bool       `json:"loadBalancerEnabled,omitempty"`
	CableDriver                   string     `json:"cableDriver,omitempty"`
	ClientToken                   *v1.Secret `json:"clientToken,omitempty"`
	GlobalnetClusterSize          uint       `json:"globalnetClusterSize,omitempty"`
	CustomDomains                 []string   `json:"customDomains,omitempty"`
	ImageOverrideArr              []string   `json:"imageOverrideArr,omitempty"`
	HealthCheckEnable             bool       `json:"healthCheckEnable,omitempty"`
	HealthCheckInterval           uint64     `json:"healthCheckInterval,omitempty"`
	HealthCheckMaxPacketLossCount uint64     `json:"healthCheckMaxPacketLossCount,omitempty"`
	CorednsCustomConfigMap        string     `json:"corednsCustomConfigMap,omitempty"`
}

type CloudPrepareConfig struct {
	// Vendor represents cloud vendor, example: aws, gcp
	Vendor string `json:"vendor,omitempty"`
	// AWS represents aws cloud prepare setup
	AWS `json:"aws,omitempty"`
}

type AWS struct {
	// AWS credentials configuration file (default "/root/.aws/credentials")
	Credentials string `json:"credentials,omitempty"`
	// GatewayInstance represents type of gateways instance machine (default "m5n.large")
	// +optional
	// +kubebuilder:default=m5n.large
	GatewayInstance string `json:"gatewayInstance,omitempty"`

	// Gateways represents the count of worker nodes that will be used to deploy the Submariner gateway
	// component on the managed cluster.
	// +optional
	// +kubebuilder:default=1
	Gateways int `json:"gateways,omitempty"`
	// AWS infra ID
	InfraID string `json:"infraID,omitempty"`
	// AWS profile to use for credentials (default "default")
	Profile string `json:"profile,omitempty"`
	// AWS regio
	Region string `json:"region,omitempty"`
	// CommonPrepareConfig represents common setup for cloud vendor
	CommonPrepareConfig `json:"commonConfig,omitempty"`
}

type CommonPrepareConfig struct {
	// Metrics port (default 8080)
	// +optional
	// +kubebuilder:default=8080
	MetricsPort uint16 `json:"metricsPort,omitempty"`
	// NAT discovery port (default 4490)
	// +optional
	// +kubebuilder:default=4490
	NatDiscoveryPort uint16 `json:"natDiscoveryPort,omitempty"`
	// IPSec NAT traversal port (default 4500)
	// +optional
	// +kubebuilder:default=4500
	NattPort uint16 `json:"nattPort,omitempty"`
	// Internal VXLAN port (default 4800)
	// +optional
	// +kubebuilder:default=4800
	VxlanPort uint16 `json:"vxlanPort,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Fabric is the Schema for the fabrics API
type Fabric struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FabricSpec   `json:"spec,omitempty"`
	Status FabricStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FabricList contains a list of Fabric
type FabricList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Fabric `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Fabric{}, &FabricList{})
}
