package controllers

import (
	"context"
	"time"

	"github.com/oam-dev/kubevela/apis/core.oam.dev/common"
	"github.com/oam-dev/kubevela/pkg/oam/util"

	corev1beta1 "github.com/wasmCloud/wasmcloud-k8s-operator/api/v1beta1"

	oamv1beta1 "github.com/oam-dev/kubevela/apis/core.oam.dev/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	fakelatticecontroller "github.com/wasmCloud/wasmcloud-k8s-operator/fake_lattice_controller"
)

var _ = Describe("Test Create Application", func() {
	const (
		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)
	var (
		application *corev1beta1.WasmCloudApplication
	)
	BeforeEach(func() {
		application = &corev1beta1.WasmCloudApplication{
			TypeMeta: metav1.TypeMeta{
				Kind:       "App",
				APIVersion: "v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "latticecontroller",
				Namespace: "default",
			},
			Spec: oamv1beta1.ApplicationSpec{
				Components: []common.ApplicationComponent{
					{
						Name: "userinfo",
						Type: "actor",
						Properties: util.Object2RawExtension(map[string]interface{}{
							"image": "wasmcloud.azurecr.io/fake:1",
						}),
						Traits: []common.ApplicationTrait{
							{
								Type: "spreadscaler",
								Properties: util.Object2RawExtension(map[string]interface{}{
									"replicas": 4,
									"spread": []map[string]interface{}{
										{
											"name": "eastcoast",
											"requirements": map[string]interface{}{
												"zone": "us-east-1",
											},
											"weight": 80,
										},
										{
											"name": "westcoast",
											"requirements": map[string]interface{}{
												"zone": "us-west-1",
											},
											"weight": 20,
										},
									},
								}),
							},
							{
								Type: "linkdef",
								Properties: util.Object2RawExtension(map[string]interface{}{
									"target": "webcap",
									"values": map[string]interface{}{
										"port": 8080,
									},
								}),
							},
						},
					},
				},
			},
			Status: corev1beta1.WasmCloudApplicationStatus{},
		}
	})
	Context("Do", func() {
		It("Should create the application", func() {
			fakecontroller := fakelatticecontroller.SetupSubscriber()

			ctx := context.Background()
			Expect(k8sClient.Create(ctx, application)).Should(Succeed())
			app := corev1beta1.WasmCloudApplication{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Namespace: "default",
					Name:      "latticecontroller",
				}, &app)

				// It takes a few updates before the status is propagated from lattice controller.
				return err == nil && app.Status.FromLatticeController == "received"
			}, timeout, interval).Should(BeTrue())

			Expect(len(app.Spec.Components)).Should(Equal(1))
			Expect(app.Status.FromLatticeController).Should(Equal("received"))

			// Check that we actually the lattice controller about our app.
			msg := fakecontroller.SpyNextMessage()
			Expect(msg).ShouldNot(BeNil())
			Expect(msg.Subject).Should(Equal("wasmbus.alc.default.put"))

			fakecontroller.Close()
		})

		It("Should delete the application", func() {
			fakecontroller := fakelatticecontroller.SetupSubscriber()

			ctx := context.Background()
			Expect(k8sClient.Delete(ctx, application)).Should(Succeed())

			msg := fakecontroller.WaitForMessage()

			Expect(msg).ShouldNot(BeNil())
			Expect(msg.Subject).Should(Equal("wasmbus.alc.default.del"))

			fakecontroller.Close()
		})
	})
})
