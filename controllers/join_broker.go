package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	operatorv1alpha1 "github.com/DanielXLee/cluster-fabric-operator/api/v1alpha1"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/discovery/globalnet"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/discovery/network"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/broker"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/datafile"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/names"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/operator/servicediscoverycr"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/operator/submarinercr"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/operator/submarinerop"
	cmdVersion "github.com/DanielXLee/cluster-fabric-operator/controllers/version"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/versions"
	submariner "github.com/submariner-io/submariner-operator/apis/submariner/v1alpha1"
	"github.com/submariner-io/submariner-operator/pkg/subctl/cmd/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var clienttoken *v1.Secret

const (
	SubmarinerNamespace = "submariner-operator" // We currently expect everything in submariner-operator
	OperatorNamespace   = "submariner-operator"
)

var nodeLabelBackoff wait.Backoff = wait.Backoff{
	Steps:    10,
	Duration: 1 * time.Second,
	Factor:   1.2,
	Jitter:   1,
}

func (r *FabricReconciler) JoinSubmarinerCluster(instance *operatorv1alpha1.Fabric, subctlData *datafile.BrokerInfo) error {
	joinConfig := instance.Spec.JoinConfig

	if joinConfig.ClusterID == "" {
		// r.Config
		// rawConfig, err := r.Config.RawConfig()
		// // This will be fatal later, no point in continuing
		// utils.ExitOnError("Error connecting to the target cluster", err)
		// clusterName := restconfig.ClusterNameFromContext(rawConfig, contextName)
		// if clusterName != nil {
		// 	clusterID = *clusterName
		// }
		joinConfig.ClusterID = "test-todo"
	}

	if valid, err := isValidClusterID(joinConfig.ClusterID); !valid {
		klog.Errorf("Cluster ID invalid: %v", err)
	}

	// clientConfig, err := config.ClientConfig()
	// utils.ExitOnError("Error connecting to the target cluster", err)

	_, failedRequirements, err := cmdVersion.CheckRequirements(r.Config)
	// We display failed requirements even if an error occurred
	if len(failedRequirements) > 0 {
		klog.Info("The target cluster fails to meet Submariner's requirements:")
		for i := range failedRequirements {
			klog.Infof("* %s", (failedRequirements)[i])
		}
	}
	if err != nil {
		klog.Errorf("Unable to check all requirements: %v", err)
		return err
	}
	if subctlData.IsConnectivityEnabled() && joinConfig.LabelGateway {
		if err := r.HandleNodeLabels(); err != nil {
			klog.Errorf("Unable to set the gateway node up: %v", err)
			return err
		}
	}

	klog.Info("Discovering network details")
	networkDetails, err := r.GetNetworkDetails()
	if err != nil {
		klog.Errorf("Error get network details: %v", err)
	}
	serviceCIDR, serviceCIDRautoDetected, err := getServiceCIDR(joinConfig.ServiceCIDR, networkDetails)
	if err != nil {
		klog.Errorf("Error determining the service CIDR: %v", err)
	}
	clusterCIDR, clusterCIDRautoDetected, err := getPodCIDR(joinConfig.ClusterCIDR, networkDetails)
	if err != nil {
		klog.Errorf("Error determining the pod CIDR: %v", err)
		return err
	}

	// brokerAdminConfig, err := subctlData.GetBrokerAdministratorConfig()
	// if err != nil {
	// 	klog.Errorf("Error retrieving broker admin config: %v", err)
	// 	return err
	// }
	// brokerAdminClientset, err := kubernetes.NewForConfig(brokerAdminConfig)
	// if err != nil {
	// 	klog.Errorf("Error retrieving broker admin connection: %v", err)
	// 	return err
	// }
	brokerCluster, err := subctlData.GetBrokerAdministratorCluster()
	if err != nil {
		klog.Errorf("unable to get broker client: %v", err)
	}
	brokerNamespace := string(subctlData.ClientToken.Data["namespace"])

	netconfig := globalnet.Config{ClusterID: joinConfig.ClusterID,
		GlobalnetCIDR:           joinConfig.GlobalnetCIDR,
		ServiceCIDR:             serviceCIDR,
		ServiceCIDRAutoDetected: serviceCIDRautoDetected,
		ClusterCIDR:             clusterCIDR,
		ClusterCIDRAutoDetected: clusterCIDRautoDetected,
		GlobalnetClusterSize:    joinConfig.GlobalnetClusterSize,
	}
	if joinConfig.GlobalnetEnabled {
		if err = r.AllocateAndUpdateGlobalCIDRConfigMap(brokerCluster.GetClient(), brokerCluster.GetAPIReader(), instance, brokerNamespace, &netconfig); err != nil {
			klog.Errorf("Error Discovering multi cluster details: %v", err)
			return err
		}
	}

	klog.Info("Deploying the Submariner operator")
	if err = submarinerop.Ensure(r.Client, r.Config, true); err != nil {
		klog.Errorf("Error deploying the operator: %v", err)
		return err
	}
	klog.Info("Creating SA for cluster")
	clienttoken, err = broker.CreateSAForCluster(brokerCluster.GetClient(), brokerCluster.GetAPIReader(), joinConfig.ClusterID)
	if err != nil {
		klog.Errorf("Error creating SA for cluster: %v", err)
		return err
	}
	if subctlData.IsConnectivityEnabled() {
		klog.Info("Deploying Submariner")
		err = submarinercr.Ensure(r.Client, OperatorNamespace, populateSubmarinerSpec(instance, subctlData, netconfig))
		if err == nil {
			klog.Info("Submariner is up and running")
		} else {
			klog.Errorf("Submariner deployment failed: %v", err)
		}
	} else if subctlData.IsServiceDiscoveryEnabled() {
		klog.Info("Deploying service discovery only")
		err = servicediscoverycr.Ensure(r.Client, OperatorNamespace, populateServiceDiscoverySpec(instance, subctlData))
		if err == nil {
			klog.Info("Service discovery is up and running")
		} else {
			klog.Errorf("Service discovery deployment failed: %v", err)
		}
	}
	return nil
}

func (r *FabricReconciler) AllocateAndUpdateGlobalCIDRConfigMap(c client.Client, reader client.Reader, instance *operatorv1alpha1.Fabric, brokerNamespace string,
	netconfig *globalnet.Config) error {
	joinConfig := instance.Spec.JoinConfig
	klog.Info("Discovering multi cluster details")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		globalnetInfo, globalnetConfigMap, err := globalnet.GetGlobalNetworks(reader, brokerNamespace)
		if err != nil {
			return fmt.Errorf("error reading Global network details on Broker: %s", err)
		}

		netconfig.GlobalnetCIDR, err = globalnet.ValidateGlobalnetConfiguration(globalnetInfo, *netconfig)
		if err != nil {
			return fmt.Errorf("error validating Globalnet configuration: %s", err)
		}

		if globalnetInfo.GlobalnetEnabled {
			netconfig.GlobalnetCIDR, err = globalnet.AssignGlobalnetIPs(globalnetInfo, *netconfig)
			if err != nil {
				return fmt.Errorf("error assigning Globalnet IPs: %s", err)
			}

			if globalnetInfo.GlobalCidrInfo[joinConfig.ClusterID] == nil ||
				globalnetInfo.GlobalCidrInfo[joinConfig.ClusterID].GlobalCIDRs[0] != netconfig.GlobalnetCIDR {
				var newClusterInfo broker.ClusterInfo
				newClusterInfo.ClusterID = joinConfig.ClusterID
				newClusterInfo.GlobalCidr = []string{netconfig.GlobalnetCIDR}

				err = broker.UpdateGlobalnetConfigMap(c, brokerNamespace, globalnetConfigMap, newClusterInfo)
				return err
			}
		}
		return err
	})
	return retryErr
}

func (r *FabricReconciler) GetNetworkDetails() (*network.ClusterNetwork, error) {
	dynClient, err := dynamic.NewForConfig(r.Config)
	if err != nil {
		return nil, err
	}

	networkDetails, err := network.Discover(dynClient, r.Client, OperatorNamespace)
	if err != nil {
		klog.Errorf("Error trying to discover network details: %v", err)
	} else if networkDetails != nil {
		networkDetails.Show()
	}
	return networkDetails, nil
}

func getPodCIDR(clusterCIDR string, nd *network.ClusterNetwork) (cidrType string, autodetected bool, err error) {
	if clusterCIDR != "" {
		if nd != nil && len(nd.PodCIDRs) > 0 && nd.PodCIDRs[0] != clusterCIDR {
			klog.Warningf("Your provided cluster CIDR for the pods (%s) does not match discovered (%s)",
				clusterCIDR, nd.PodCIDRs[0])
		}
		return clusterCIDR, false, nil
	} else if nd != nil && len(nd.PodCIDRs) > 0 {
		return nd.PodCIDRs[0], true, nil
	} else {
		cidrType, err = askForCIDR("Pod")
		return cidrType, false, err
	}
}

func getServiceCIDR(serviceCIDR string, nd *network.ClusterNetwork) (cidrType string, autodetected bool, err error) {
	if serviceCIDR != "" {
		if nd != nil && len(nd.ServiceCIDRs) > 0 && nd.ServiceCIDRs[0] != serviceCIDR {
			klog.Warningf("Your provided service CIDR (%s) does not match discovered (%s)",
				serviceCIDR, nd.ServiceCIDRs[0])
		}
		return serviceCIDR, false, nil
	} else if nd != nil && len(nd.ServiceCIDRs) > 0 {
		return nd.ServiceCIDRs[0], true, nil
	} else {
		cidrType, err = askForCIDR("ClusterIP service")
		return cidrType, false, err
	}
}

func askForCIDR(name string) (string, error) {
	var qs = []*survey.Question{{
		Name:     "cidr",
		Prompt:   &survey.Input{Message: fmt.Sprintf("What's the %s CIDR for your cluster?", name)},
		Validate: survey.Required,
	}}

	answers := struct {
		Cidr string
	}{}

	err := survey.Ask(qs, &answers)
	if err != nil {
		return "", err
	} else {
		return strings.TrimSpace(answers.Cidr), nil
	}
}

func isValidClusterID(clusterID string) (bool, error) {
	// Make sure the clusterid is a valid DNS-1123 string
	if match, _ := regexp.MatchString("^[a-z0-9][a-z0-9.-]*[a-z0-9]$", clusterID); !match {
		return false, fmt.Errorf("cluster IDs must be valid DNS-1123 names, with only lowercase alphanumerics,\n"+
			"'.' or '-' (and the first and last characters must be alphanumerics).\n"+
			"%s doesn't meet these requirements", clusterID)
	}
	return true, nil
}

func populateSubmarinerSpec(instance *operatorv1alpha1.Fabric, subctlData *datafile.BrokerInfo, netconfig globalnet.Config) submariner.SubmarinerSpec {
	joinConfig := instance.Spec.JoinConfig
	brokerURL := subctlData.BrokerURL
	if idx := strings.Index(brokerURL, "://"); idx >= 0 {
		// Submariner doesn't work with a schema prefix
		brokerURL = brokerURL[(idx + 3):]
	}

	// if our network discovery code was capable of discovering those CIDRs
	// we don't need to explicitly set it in the operator
	crServiceCIDR := ""
	if !netconfig.ServiceCIDRAutoDetected {
		crServiceCIDR = netconfig.ServiceCIDR
	}

	crClusterCIDR := ""
	if !netconfig.ClusterCIDRAutoDetected {
		crClusterCIDR = netconfig.ClusterCIDR
	}
	// customDomains := ""
	if joinConfig.CustomDomains == nil && subctlData.CustomDomains != nil {
		joinConfig.CustomDomains = *subctlData.CustomDomains
	}

	submarinerSpec := submariner.SubmarinerSpec{
		Repository:               getImageRepo(instance),
		Version:                  getImageVersion(instance),
		CeIPSecNATTPort:          joinConfig.NattPort,
		CeIPSecIKEPort:           joinConfig.IkePort,
		CeIPSecDebug:             joinConfig.IpsecDebug,
		CeIPSecForceUDPEncaps:    joinConfig.ForceUDPEncaps,
		CeIPSecPreferredServer:   joinConfig.PreferredServer,
		CeIPSecPSK:               base64.StdEncoding.EncodeToString(subctlData.IPSecPSK.Data["psk"]),
		BrokerK8sCA:              base64.StdEncoding.EncodeToString(subctlData.ClientToken.Data["ca.crt"]),
		BrokerK8sRemoteNamespace: string(subctlData.ClientToken.Data["namespace"]),
		BrokerK8sApiServerToken:  string(clienttoken.Data["token"]),
		BrokerK8sApiServer:       brokerURL,
		Broker:                   "k8s",
		NatEnabled:               joinConfig.NatTraversal,
		Debug:                    joinConfig.SubmarinerDebug,
		ClusterID:                joinConfig.ClusterID,
		ServiceCIDR:              crServiceCIDR,
		ClusterCIDR:              crClusterCIDR,
		Namespace:                SubmarinerNamespace,
		CableDriver:              joinConfig.CableDriver,
		ServiceDiscoveryEnabled:  subctlData.IsServiceDiscoveryEnabled(),
		ImageOverrides:           getImageOverrides(instance),
	}
	if netconfig.GlobalnetCIDR != "" {
		submarinerSpec.GlobalCIDR = netconfig.GlobalnetCIDR
	}
	if joinConfig.CorednsCustomConfigMap != "" {
		namespace, name := getCustomCoreDNSParams(instance)
		submarinerSpec.CoreDNSCustomConfig = &submariner.CoreDNSCustomConfig{
			ConfigMapName: name,
			Namespace:     namespace,
		}
	}
	if subctlData.CustomDomains != nil && len(*subctlData.CustomDomains) > 0 {
		submarinerSpec.CustomDomains = *subctlData.CustomDomains
	}
	return submarinerSpec
}

func getImageVersion(instance *operatorv1alpha1.Fabric) string {
	version := instance.Spec.JoinConfig.ImageVersion

	if version == "" {
		version = versions.DefaultSubmarinerOperatorVersion
	}

	return version
}

func getImageRepo(instance *operatorv1alpha1.Fabric) string {
	repo := instance.Spec.JoinConfig.Repository

	if repo == "" {
		repo = versions.DefaultRepo
	}

	return repo
}

func removeSchemaPrefix(brokerURL string) string {
	if idx := strings.Index(brokerURL, "://"); idx >= 0 {
		// Submariner doesn't work with a schema prefix
		brokerURL = brokerURL[(idx + 3):]
	}

	return brokerURL
}

func populateServiceDiscoverySpec(instance *operatorv1alpha1.Fabric, subctlData *datafile.BrokerInfo) *submariner.ServiceDiscoverySpec {
	brokerURL := removeSchemaPrefix(subctlData.BrokerURL)
	joinConfig := instance.Spec.JoinConfig
	var customDomains []string
	if joinConfig.CustomDomains == nil && subctlData.CustomDomains != nil {
		customDomains = *subctlData.CustomDomains
	}

	serviceDiscoverySpec := submariner.ServiceDiscoverySpec{
		Repository:               joinConfig.Repository,
		Version:                  joinConfig.ImageVersion,
		BrokerK8sCA:              base64.StdEncoding.EncodeToString(subctlData.ClientToken.Data["ca.crt"]),
		BrokerK8sRemoteNamespace: string(subctlData.ClientToken.Data["namespace"]),
		BrokerK8sApiServerToken:  string(clienttoken.Data["token"]),
		BrokerK8sApiServer:       brokerURL,
		Debug:                    joinConfig.SubmarinerDebug,
		ClusterID:                joinConfig.ClusterID,
		Namespace:                SubmarinerNamespace,
		ImageOverrides:           getImageOverrides(instance),
	}

	if joinConfig.CorednsCustomConfigMap != "" {
		namespace, name := getCustomCoreDNSParams(instance)
		serviceDiscoverySpec.CoreDNSCustomConfig = &submariner.CoreDNSCustomConfig{
			ConfigMapName: name,
			Namespace:     namespace,
		}
	}

	if len(customDomains) > 0 {
		serviceDiscoverySpec.CustomDomains = customDomains
	}
	return &serviceDiscoverySpec
}

// func operatorImage() string {
// 	version := imageVersion
// 	repo := repository

// 	if imageVersion == "" {
// 		version = versions.DefaultSubmarinerOperatorVersion
// 	}

// 	if repository == "" {
// 		repo = versions.DefaultRepo
// 	}

// 	return images.GetImagePath(repo, version, names.OperatorImage, names.OperatorComponent, getImageOverrides())
// }

func getImageOverrides(instance *operatorv1alpha1.Fabric) map[string]string {
	joinConfig := instance.Spec.JoinConfig
	if len(joinConfig.ImageOverrideArr) > 0 {
		imageOverrides := make(map[string]string)
		for _, s := range joinConfig.ImageOverrideArr {
			key := strings.Split(s, "=")[0]
			if invalidImageName(key) {
				utils.ExitWithErrorMsg(fmt.Sprintf("Invalid image name %s provided. Please choose from %q", key, names.ValidImageNames))
			}
			value := strings.Split(s, "=")[1]
			imageOverrides[key] = value
		}
		return imageOverrides
	}
	return nil
}

func invalidImageName(key string) bool {
	for _, name := range names.ValidImageNames {
		if key == name {
			return false
		}
	}
	return true
}

func isValidCustomCoreDNSConfig(instance *operatorv1alpha1.Fabric) error {
	corednsCustomConfigMap := instance.Spec.JoinConfig.CorednsCustomConfigMap
	if corednsCustomConfigMap != "" && strings.Count(corednsCustomConfigMap, "/") > 1 {
		return fmt.Errorf("coredns-custom-configmap should be in <namespace>/<name> format, namespace is optional")
	}
	return nil
}

func getCustomCoreDNSParams(instance *operatorv1alpha1.Fabric) (namespace, name string) {
	corednsCustomConfigMap := instance.Spec.JoinConfig.CorednsCustomConfigMap
	if corednsCustomConfigMap != "" {
		name = corednsCustomConfigMap
		paramList := strings.Split(corednsCustomConfigMap, "/")
		if len(paramList) > 1 {
			namespace = paramList[0]
			name = paramList[1]
		}
	}
	return namespace, name
}

func (r *FabricReconciler) HandleNodeLabels() error {
	// _, clientset, err := restconfig.Clients(config)
	// utils.ExitOnError("Unable to set the Kubernetes cluster connection up", err)
	// List Submariner-labeled nodes
	const submarinerGatewayLabel = "submariner.io/gateway"
	const trueLabel = "true"
	selector, err := labels.Parse("submariner.io/gateway=true")
	if err != nil {
		return err
	}
	opts := &client.ListOptions{
		LabelSelector: selector,
	}
	nodes := &v1.NodeList{}
	if err := r.Client.List(context.TODO(), nodes, opts); err != nil {
		return err
	}
	// labeledNodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	// if err != nil {
	// 	return err
	// }
	if len(nodes.Items) > 0 {
		klog.Infof("* There are %d labeled nodes in the cluster:", len(nodes.Items))
		for _, node := range nodes.Items {
			klog.Infof("  - %s", node.GetName())
		}
	} else {
		answer, err := r.askForGatewayNode()
		if err != nil {
			return err
		}
		if answer.Node == "" {
			klog.Info("* No worker node found to label as the gateway")
		} else {
			err = r.addLabelsToNode(answer.Node, map[string]string{submarinerGatewayLabel: trueLabel})
			utils.ExitOnError("Error labeling the gateway node", err)
		}
	}
	return nil
}
func (r *FabricReconciler) askForGatewayNode() (struct{ Node string }, error) {
	// List the worker nodes and select one
	workerNodes := &v1.NodeList{}
	workerSelector, err := labels.Parse("submariner.io/gateway=true")
	if err != nil {
		return struct{ Node string }{}, err
	}
	workerOpts := &client.ListOptions{
		LabelSelector: workerSelector,
	}

	// workerNodes, err := clientset.CoreV1().Nodes().List(
	// 	context.TODO(), metav1.ListOptions{LabelSelector: "node-role.kubernetes.io/worker"})
	if err := r.Client.List(context.TODO(), workerNodes, workerOpts); err != nil {
		return struct{ Node string }{}, err
	}
	if len(workerNodes.Items) == 0 {
		// In some deployments (like KIND), worker nodes are not explicitly labelled. So list non-master nodes.
		// workerNodes, err = clientset.CoreV1().Nodes().List(
		// 	context.TODO(), metav1.ListOptions{LabelSelector: "!node-role.kubernetes.io/master"})
		// if err != nil {
		// 	return struct{ Node string }{}, err
		// }

		workerNodes := &v1.NodeList{}
		workerSelector, err := labels.Parse("!node-role.kubernetes.io/master")
		if err != nil {
			return struct{ Node string }{}, err
		}
		workerOpts := &client.ListOptions{
			LabelSelector: workerSelector,
		}

		// workerNodes, err := clientset.CoreV1().Nodes().List(
		// 	context.TODO(), metav1.ListOptions{LabelSelector: "node-role.kubernetes.io/worker"})
		if err := r.Client.List(context.TODO(), workerNodes, workerOpts); err != nil {
			return struct{ Node string }{}, err
		}
		if len(workerNodes.Items) == 0 {
			return struct{ Node string }{}, nil
		}
	}
	// Return the first node
	return struct{ Node string }{workerNodes.Items[0].GetName()}, nil
}

// this function was sourced from:
// https://github.com/kubernetes/kubernetes/blob/a3ccea9d8743f2ff82e41b6c2af6dc2c41dc7b10/test/utils/density_utils.go#L36
func (r *FabricReconciler) addLabelsToNode(nodeName string, labelsToAdd map[string]string) error {
	var tokens = make([]string, 0, len(labelsToAdd))
	for k, v := range labelsToAdd {
		tokens = append(tokens, fmt.Sprintf("\"%s\":\"%s\"", k, v))
	}

	labelString := "{" + strings.Join(tokens, ",") + "}"
	// patch := fmt.Sprintf(`{"metadata":{"labels":%v}}`, labelString)
	patch := []byte(fmt.Sprintf(`{"metadata":{"labels":%v}}`, labelString))
	// retry is necessary because nodes get updated every 10 seconds, and a patch can happen
	// in the middle of an update

	var lastErr error
	err := wait.ExponentialBackoff(nodeLabelBackoff, func() (bool, error) {
		node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: nodeName}}
		lastErr := r.Client.Patch(context.TODO(), node, client.RawPatch(types.StrategicMergePatchType, patch))
		// _, lastErr = c.CoreV1().Nodes().Patch(context.TODO(), nodeName, types.MergePatchType, []byte(patch), metav1.PatchOptions{})
		if lastErr != nil {
			if !errors.IsConflict(lastErr) {
				return false, lastErr
			}
			return false, nil
		} else {
			return true, nil
		}
	})

	if err == wait.ErrWaitTimeout {
		return lastErr
	}

	return err
}
