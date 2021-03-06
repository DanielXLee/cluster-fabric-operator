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

package namespace

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Ensure functions updates or installs the operator CRDs in the cluster
func Ensure(c client.Client, namespace string) (bool, error) {
	// clientSet, err := clientset.NewForConfig(restConfig)
	// if err != nil {
	// 	return false, err
	// }

	ns := &v1.Namespace{}
	nsKey := types.NamespacedName{Name: namespace}
	err := c.Get(context.TODO(), nsKey, ns)
	// _, err = clientSet.CoreV1().Namespaces().Create(ns)

	if err == nil {
		return true, nil
	} else if errors.IsAlreadyExists(err) {
		return false, nil
	} else {
		return false, err
	}
}
