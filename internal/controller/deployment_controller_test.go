/*
Copyright 2024.

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

package controller

import (
	"context"

	hcv2 "github.com/fluxcd/helm-controller/api/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	hmc "github.com/Mirantis/hmc/api/v1alpha1"
)

var _ = Describe("Deployment Controller", func() {
	Context("When reconciling a resource", func() {
		const deploymentName = "test-deployment"
		const templateName = "test-template"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      deploymentName,
			Namespace: "default",
		}
		deployment := &hmc.Deployment{}
		template := &hmc.Template{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Template")
			err := k8sClient.Get(ctx, typeNamespacedName, template)
			if err != nil && errors.IsNotFound(err) {
				template := &hmc.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name:      templateName,
						Namespace: "default",
					},
					Spec: hmc.TemplateSpec{
						Provider: "aws",
						Helm: hmc.HelmSpec{
							ChartRef: &hcv2.CrossNamespaceSourceReference{
								Kind:      "HelmChart",
								Name:      "ref-test",
								Namespace: "default",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, template)).To(Succeed())
				template.Status = hmc.TemplateStatus{
					TemplateValidationStatus: hmc.TemplateValidationStatus{
						Valid: true,
					},
					Config: &apiextensionsv1.JSON{
						Raw: []byte(`{"foo":"bar"}`),
					},
				}
				Expect(k8sClient.Status().Update(ctx, template)).To(Succeed())
			}
			By("creating the custom resource for the Kind Deployment")
			err = k8sClient.Get(ctx, typeNamespacedName, deployment)
			if err != nil && errors.IsNotFound(err) {
				deployment := &hmc.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      deploymentName,
						Namespace: "default",
					},
					Spec: hmc.DeploymentSpec{
						Template: templateName,
					},
				}
				Expect(k8sClient.Create(ctx, deployment)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &hmc.Deployment{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Deployment")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &DeploymentReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})