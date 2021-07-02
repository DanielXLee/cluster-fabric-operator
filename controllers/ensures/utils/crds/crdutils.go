/*
© 2021 Red Hat, Inc. and others.

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

package crdutils

import (
	"context"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type CRDUpdater interface {
	Create(ctx context.Context, customResourceDefinition *apiextensionsv1.CustomResourceDefinition, opts metav1.CreateOptions) (*apiextensionsv1.CustomResourceDefinition, error)
	Update(ctx context.Context, customResourceDefinition *apiextensionsv1.CustomResourceDefinition, opts metav1.UpdateOptions) (*apiextensionsv1.CustomResourceDefinition, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiextensionsv1.CustomResourceDefinition, error)
}

func NewFromRestConfig(config *rest.Config) (CRDUpdater, error) {
	apiext, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating the api extensions client: %s", err)
	}
	return NewFromClientSet(apiext), nil
}

func NewFromClientSet(cs clientset.Interface) CRDUpdater {
	return cs.ApiextensionsV1().CustomResourceDefinitions()
}
