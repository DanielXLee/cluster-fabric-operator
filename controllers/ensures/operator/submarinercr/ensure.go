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

package submarinercr

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	submariner "github.com/submariner-io/submariner-operator/apis/submariner/v1alpha1"
)

const (
	SubmarinerName = "submariner"
)

var backOff wait.Backoff = wait.Backoff{
	Steps:    10,
	Duration: 500 * time.Millisecond,
	Factor:   1.5,
	Cap:      20 * time.Second,
}

func Ensure(c client.Client, namespace string, submarinerSpec submariner.SubmarinerSpec) error {
	submarinerCR := &submariner.Submariner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SubmarinerName,
			Namespace: namespace,
		},
		Spec: submarinerSpec,
	}

	return wait.ExponentialBackoff(backOff, func() (bool, error) {
		if err := c.Create(context.TODO(), submarinerCR); !errors.IsAlreadyExists(err) {
			return true, err
		}

		fg := metav1.DeletePropagationForeground
		delOpts := &client.DeleteOptions{PropagationPolicy: &fg}
		err := c.Delete(context.TODO(), submarinerCR, delOpts)

		return false, client.IgnoreNotFound(err)
	})
}
