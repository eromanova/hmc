// Copyright 2024
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package credspropagation

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hmc "github.com/Mirantis/hmc/api/v1alpha1"
)

type PropagationCfg struct {
	Client          client.Client
	ManagedCluster  *hmc.ManagedCluster
	KubeconfSecret  *corev1.Secret
	SystemNamespace string
}

func applyCCMConfigs(ctx context.Context, kubeconfSecret *corev1.Secret, objects ...client.Object) error {
	clnt, err := makeClientFromSecret(kubeconfSecret)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}
	for _, object := range objects {
		if err := clnt.Patch(
			ctx,
			object,
			client.Apply,
			client.FieldOwner("hmc-controller"),
		); err != nil {
			return fmt.Errorf("failed to apply CCM config object %s: %w", object.GetName(), err)
		}
	}
	return nil
}

func makeSecret(name, namespace string, data map[string][]byte) *corev1.Secret {
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
	s.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))
	return s
}

func makeConfigMap(name, namespace string, data map[string]string) *corev1.ConfigMap {
	c := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
	c.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))
	return c
}

func makeClientFromSecret(kubeconfSecret *corev1.Secret) (client.Client, error) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfSecret.Data["value"])
	if err != nil {
		return nil, err
	}
	cl, err := client.New(restConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}
	return cl, nil
}
