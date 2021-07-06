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

package crds

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/operator/common/embeddedyamls"
	"github.com/DanielXLee/cluster-fabric-operator/controllers/ensures/utils"
)

func Ensure(c client.Client) (bool, error) {
	return utils.CreateOrUpdateEmbeddedCRD(c, embeddedyamls.Manifests_deploy_crds_submariner_io_servicediscoveries_yaml)
}
