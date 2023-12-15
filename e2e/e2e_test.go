package e2e

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/k8snetworkplumbingwg/multus-dynamic-networks-controller/e2e/client"
	"github.com/k8snetworkplumbingwg/multus-dynamic-networks-controller/e2e/status"
	nettypes "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
)

func TestDynamicNetworksControllerE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multus Dynamic Networks Controller")
}

var _ = Describe("Multus dynamic networks controller", func() {
	const (
		namespace   = "ns1"
		networkName = "tenant-network"
		podName     = "tiny-winy-pod"
		timeout     = 15 * time.Second
	)
	var clients *client.E2EClient

	BeforeEach(func() {
		config, err := clusterConfig()
		Expect(err).NotTo(HaveOccurred())

		clients, err = client.New(config)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("a simple network-attachment-definition", func() {
		const initialPodIfaceName = "net1"

		BeforeEach(func() {
			_, err := clients.AddNamespace(namespace)
			Expect(err).NotTo(HaveOccurred())
			_, err = clients.AddNetAttachDef(macvlanNetworkWithoutIPAM(networkName, namespace, lowerDeviceName()))
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(clients.DeleteNamespace(namespace)).To(Succeed())
		})

		// filterPodNonDefaultNetworks := func() []nettypes.NetworkStatus {
		// 	return status.FilterPodsNetworkStatus(clients, namespace, podName, func(networkStatus nettypes.NetworkStatus) bool {
		// 		return !networkStatus.Default
		// 	})
		// }

		isTestNamespaceEmpty := func() bool {
			pods, err := clients.ListPods(namespace, appLabel(podName))
			if err != nil {
				return false
			}
			return len(pods.Items) == 0
		}

		// Context("a provisioned pod having network selection elements", func() {
		// 	var pod *corev1.Pod

		// 	initialIfaceNetworkStatus := nettypes.NetworkStatus{
		// 		Name:      namespacedName(namespace, networkName),
		// 		Interface: initialPodIfaceName,
		// 	}

		// 	BeforeEach(func() {
		// 		var err error
		// 		pod, err = clients.ProvisionPod(
		// 			podName,
		// 			namespace,
		// 			podAppLabel(podName),
		// 			PodNetworkSelectionElements(
		// 				dynamicNetworkInfo{
		// 					namespace:   namespace,
		// 					networkName: networkName,
		// 					ifaceName:   initialPodIfaceName,
		// 				}),
		// 		)
		// 		Expect(err).NotTo(HaveOccurred())
		// 		Expect(filterPodNonDefaultNetworks()).Should(
		// 			WithTransform(
		// 				status.CleanMACAddressesFromStatus(),
		// 				ConsistOf(initialIfaceNetworkStatus),
		// 			))
		// 	})

		// 	AfterEach(func() {
		// 		Expect(clients.DeletePod(pod)).To(Succeed())
		// 		Eventually(isTestNamespaceEmpty, timeout).Should(BeTrue())
		// 	})

		// 	It("manages to add a new interface to a running pod", func() {
		// 		const ifaceToAdd = "ens58"

		// 		Expect(clients.AddNetworkToPod(pod, &nettypes.NetworkSelectionElement{
		// 			Name:             networkName,
		// 			Namespace:        namespace,
		// 			InterfaceRequest: ifaceToAdd,
		// 		})).To(Succeed())
		// 		Eventually(filterPodNonDefaultNetworks, timeout).Should(
		// 			WithTransform(
		// 				status.CleanMACAddressesFromStatus(),
		// 				ConsistOf(
		// 					nettypes.NetworkStatus{
		// 						Name:      namespacedName(namespace, networkName),
		// 						Interface: ifaceToAdd,
		// 					},
		// 					initialIfaceNetworkStatus)))
		// 	})

		// 	It("manages to remove an interface from a running pod", func() {
		// 		const ifaceToRemove = initialPodIfaceName

		// 		Expect(clients.RemoveNetworkFromPod(pod, networkName, namespace, ifaceToRemove)).To(Succeed())
		// 		Eventually(filterPodNonDefaultNetworks, timeout).Should(BeEmpty())
		// 	})

		// 	It("interfaces can be added / removed in the same operation", func() {
		// 		const newIface = "eth321"
		// 		Expect(clients.SetPodNetworks(
		// 			pod,
		// 			NetworkSelectionElements(
		// 				dynamicNetworkInfo{
		// 					namespace:   namespace,
		// 					networkName: networkName,
		// 					ifaceName:   newIface,
		// 				})...,
		// 		)).To(Succeed())
		// 		Eventually(filterPodNonDefaultNetworks, timeout).Should(
		// 			WithTransform(
		// 				status.CleanMACAddressesFromStatus(),
		// 				ConsistOf(
		// 					nettypes.NetworkStatus{
		// 						Name:      namespacedName(namespace, networkName),
		// 						Interface: newIface,
		// 					},
		// 				)))
		// 	})

		// 	Context("a network with IPAM", func() {
		// 		const (
		// 			ifaceToAddWithIPAM = "ens202"
		// 			ipAddressToAdd     = "10.10.10.111"
		// 			ipamNetworkToAdd   = "tenant-network-ipam"
		// 			netmaskLen         = 24
		// 		)

		// 		var (
		// 			desiredMACAddr string
		// 		)

		// 		BeforeEach(func() {
		// 			mac, err := generateMacAddress()
		// 			Expect(err).NotTo(HaveOccurred())
		// 			desiredMACAddr = mac.String()

		// 			_, err = clients.AddNetAttachDef(macvlanNetworkWitStaticIPAM(ipamNetworkToAdd, namespace, lowerDeviceName()))
		// 			Expect(err).NotTo(HaveOccurred())
		// 			Expect(clients.AddNetworkToPod(pod, &nettypes.NetworkSelectionElement{
		// 				Name:             ipamNetworkToAdd,
		// 				Namespace:        namespace,
		// 				IPRequest:        []string{ipWithMask(ipAddressToAdd, netmaskLen)},
		// 				InterfaceRequest: ifaceToAddWithIPAM,
		// 				MacRequest:       desiredMACAddr,
		// 			})).To(Succeed())
		// 		})

		// 		It("can be hotplugged into a running pod", func() {
		// 			Eventually(filterPodNonDefaultNetworks, timeout).Should(
		// 				ContainElements(
		// 					nettypes.NetworkStatus{
		// 						Name:      namespacedName(namespace, ipamNetworkToAdd),
		// 						Interface: ifaceToAddWithIPAM,
		// 						IPs:       []string{ipAddressToAdd},
		// 						Mac:       desiredMACAddr,
		// 					},
		// 				))
		// 		})

		// 		It("can be hot unplugged from a running pod", func() {
		// 			const ifaceToRemove = ifaceToAddWithIPAM
		// 			pods, err := clients.ListPods(namespace, appLabel(podName))
		// 			Expect(err).NotTo(HaveOccurred())
		// 			pod = &pods.Items[0]

		// 			Expect(clients.RemoveNetworkFromPod(pod, networkName, namespace, ifaceToRemove)).To(Succeed())
		// 			Eventually(filterPodNonDefaultNetworks, timeout).Should(
		// 				Not(ContainElements(
		// 					nettypes.NetworkStatus{
		// 						Name:      namespacedName(namespace, ipamNetworkToAdd),
		// 						Interface: ifaceToAddWithIPAM,
		// 						IPs:       []string{ipAddressToAdd},
		// 						Mac:       desiredMACAddr,
		// 					},
		// 				)))
		// 		})
		// 	})
		// })

		// Context("a provisioned pod featuring *only* the cluster's default network", func() {
		// 	var (
		// 		pod            *corev1.Pod
		// 		desiredMACAddr string
		// 	)

		// 	BeforeEach(func() {
		// 		mac, err := generateMacAddress()
		// 		Expect(err).NotTo(HaveOccurred())
		// 		desiredMACAddr = mac.String()

		// 		pod, err = clients.ProvisionPod(
		// 			podName,
		// 			namespace,
		// 			podAppLabel(podName),
		// 			PodNetworkSelectionElements())
		// 		Expect(err).NotTo(HaveOccurred())
		// 	})

		// 	AfterEach(func() {
		// 		Expect(clients.DeletePod(pod)).To(Succeed())
		// 		Eventually(isTestNamespaceEmpty, timeout).Should(BeTrue())
		// 	})

		// 	It("manages to add a new interface to a running pod", func() {
		// 		const (
		// 			ifaceToAdd = "ens58"
		// 		)

		// 		Expect(clients.AddNetworkToPod(pod, &nettypes.NetworkSelectionElement{
		// 			Name:             networkName,
		// 			Namespace:        namespace,
		// 			InterfaceRequest: ifaceToAdd,
		// 			MacRequest:       desiredMACAddr,
		// 		})).To(Succeed())
		// 		Eventually(filterPodNonDefaultNetworks, timeout).Should(
		// 			ConsistOf(
		// 				nettypes.NetworkStatus{
		// 					Name:      namespacedName(namespace, networkName),
		// 					Interface: ifaceToAdd,
		// 					Mac:       desiredMACAddr,
		// 				}))
		// 	})
		// })

		Context("a provisioned pod whose network selection elements do not feature the interface name", func() {
			const (
				ifaceToAdd = "ens58"
			)

			var (
				pod                      *corev1.Pod
				initialPodsNetworkStatus []nettypes.NetworkStatus
				desiredMACAddr           string
			)

			runtimePodNetworkStatus := func() []nettypes.NetworkStatus {
				return status.FilterPodsNetworkStatus(
					clients,
					namespace,
					podName,
					func(networkStatus nettypes.NetworkStatus) bool {
						return true
					},
				)
			}

			BeforeEach(func() {
				mac, err := generateMacAddress()
				Expect(err).NotTo(HaveOccurred())
				desiredMACAddr = mac.String()

				cmd0 := exec.Command("bash", "-c", "kubectl get pods tiny-winy-pod -n ns1 -o yaml | head -n 35")
				var stdout0 bytes.Buffer
				cmd0.Stdout = &stdout0
				cmd0.Run()
				fmt.Println(stdout0.String())

				cmdExec0 := exec.Command("bash", "-c", "kubectl exec tiny-winy-pod -n ns1 -- ip a")
				var stdoutExec0 bytes.Buffer
				cmdExec0.Stdout = &stdoutExec0
				cmdExec0.Run()
				fmt.Println(stdoutExec0.String())

				pod, err = clients.ProvisionPod(
					podName,
					namespace,
					podAppLabel(podName),
					PodNetworkSelectionElements(
						dynamicNetworkInfo{
							namespace:   namespace,
							networkName: networkName,
						}),
				)
				Expect(err).NotTo(HaveOccurred())

				initialPodsNetworkStatus = clients.NetworkStatus(pod)
				const defaultNetworkPlusInitialAttachment = 2
				Expect(initialPodsNetworkStatus).To(HaveLen(defaultNetworkPlusInitialAttachment))

				fmt.Println("-------------------------------------------------------------")
				fmt.Println("-------------------------------------------------------------")
				fmt.Println("-------------------------------------------------------------")

				cmd1 := exec.Command("bash", "-c", "kubectl get pods tiny-winy-pod -n ns1 -o yaml | head -n 35")
				var stdout1 bytes.Buffer
				cmd1.Stdout = &stdout1
				cmd1.Run()
				fmt.Println(stdout1.String())

				cmdExec1 := exec.Command("bash", "-c", "kubectl exec tiny-winy-pod -n ns1 -- ip a")
				var stdoutExec1 bytes.Buffer
				cmdExec1.Stdout = &stdoutExec1
				cmdExec1.Run()
				fmt.Println(stdoutExec1.String())

				Expect(clients.AddNetworkToPod(pod, &nettypes.NetworkSelectionElement{
					Name:             networkName,
					Namespace:        namespace,
					InterfaceRequest: ifaceToAdd,
					MacRequest:       desiredMACAddr,
				})).To(Succeed())

				fmt.Println("-------------------------------------------------------------")
				fmt.Println("-------------------------------------------------------------")
				fmt.Println("-------------------------------------------------------------")

				// time.Sleep(2 * time.Second)

				cmd2 := exec.Command("bash", "-c", "kubectl get pods tiny-winy-pod -n ns1 -o yaml | head -n 35")
				var stdout2 bytes.Buffer
				cmd2.Stdout = &stdout2
				cmd2.Run()
				fmt.Println(stdout2.String())

				cmdExec2 := exec.Command("bash", "-c", "kubectl exec tiny-winy-pod -n ns1 -- ip a")
				var stdoutExec2 bytes.Buffer
				cmdExec2.Stdout = &stdoutExec2
				cmdExec2.Run()
				fmt.Println(stdoutExec2.String())

				fmt.Println("-------------------------------------------------------------")
				fmt.Println("-------------------------------------------------------------")
				fmt.Println("-------------------------------------------------------------")

				time.Sleep(2 * time.Second)

				cmd3 := exec.Command("bash", "-c", "kubectl get pods tiny-winy-pod -n ns1 -o yaml | head -n 35")
				var stdout3 bytes.Buffer
				cmd3.Stdout = &stdout3
				cmd3.Run()
				fmt.Println(stdout3.String())

				cmdExec3 := exec.Command("bash", "-c", "kubectl exec tiny-winy-pod -n ns1 -- ip a")
				var stdoutExec3 bytes.Buffer
				cmdExec3.Stdout = &stdoutExec3
				cmdExec3.Run()
				fmt.Println(stdoutExec3.String())

				By("the attachment is ignored since we cannot reconcile without knowing the interface name of all attachments")
				Consistently(runtimePodNetworkStatus, timeout).Should(ConsistOf(initialPodsNetworkStatus))
			})

			AfterEach(func() {
				Expect(clients.DeletePod(pod)).To(Succeed())
				Eventually(isTestNamespaceEmpty, timeout).Should(BeTrue())
			})

			runningPod := func() *corev1.Pod {
				pods, err := clients.ListPods(namespace, appLabel(podName))
				ExpectWithOffset(1, err).NotTo(HaveOccurred())
				ExpectWithOffset(1, pods.Items).NotTo(BeEmpty())
				return &pods.Items[0]
			}

			It("manages to add a new interface to a running pod once the desired state features the interface names", func() {
				Expect(true).To(BeFalse())

				By("setting the interface name in the existing attachment")
				Expect(clients.SetInterfaceNamesOnPodsNetworkSelectionElements(runningPod())).To(Succeed())

				fmt.Println("-------------------------------------------------------------")
				fmt.Println("-------------------------------------------------------------")
				fmt.Println("-------------------------------------------------------------")

				cmd3 := exec.Command("bash", "-c", "kubectl get pods tiny-winy-pod -n ns1 -o yaml | head -n 35")
				var stdout3 bytes.Buffer
				cmd3.Stdout = &stdout3
				cmd3.Run()
				fmt.Println(stdout3.String())

				Eventually(runtimePodNetworkStatus, 2*timeout).Should( // this test takes longer, for unknown reasons
					ConsistOf(
						append(initialPodsNetworkStatus, nettypes.NetworkStatus{
							Name:      namespacedName(namespace, networkName),
							Interface: ifaceToAdd,
							Mac:       desiredMACAddr,
						}),
					))
			})
		})

	})
})

// https://stackoverflow.com/questions/21018729/generate-mac-address-in-go
func generateMacAddress() (net.HardwareAddr, error) {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	buf[0] = (buf[0] | 2) & 0xfe // Set local bit, ensure unicast addres

	return buf, nil
}

func clusterConfig() (*rest.Config, error) {
	const kubeconfig = "KUBECONFIG"

	kubeconfigPath, found := os.LookupEnv(kubeconfig)
	if !found {
		homePath := os.Getenv("HOME")
		kubeconfigPath = fmt.Sprintf("%s/.kube/config", homePath)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func macvlanNetworkWithoutIPAM(networkName string, namespaceName string, lowerDevice string) *nettypes.NetworkAttachmentDefinition {
	macvlanConfig := fmt.Sprintf(`{
        "cniVersion": "0.3.0",
        "disableCheck": true,
        "plugins": [
            {
                "type": "macvlan",
                "master": "%s",
                "mode": "bridge"
            }
        ]
    }`, lowerDevice)
	return generateNetAttachDefSpec(networkName, namespaceName, macvlanConfig)
}

func macvlanNetworkWitStaticIPAM(networkName string, namespaceName string, lowerDevice string) *nettypes.NetworkAttachmentDefinition {
	macvlanConfig := fmt.Sprintf(`{
        "cniVersion": "0.3.0",
        "disableCheck": true,
        "name": "%s",
        "plugins": [
			{
				"type": "macvlan",
				"capabilities": { "ips": true },
				"master": "%s",
				"mode": "bridge",
				"ipam": {
					"type": "static"
				}
			}, {
				"type": "tuning"
			}
        ]
    }`, networkName, lowerDevice)
	return generateNetAttachDefSpec(networkName, namespaceName, macvlanConfig)
}

func generateNetAttachDefSpec(name, namespace, config string) *nettypes.NetworkAttachmentDefinition {
	return &nettypes.NetworkAttachmentDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "NetworkAttachmentDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: nettypes.NetworkAttachmentDefinitionSpec{
			Config: config,
		},
	}
}

func podAppLabel(appName string) map[string]string {
	const (
		app = "app"
	)

	return map[string]string{app: appName}
}

type dynamicNetworkInfo struct {
	namespace   string
	networkName string
	ifaceName   string
}

func PodNetworkSelectionElements(networkConfig ...dynamicNetworkInfo) map[string]string {
	podNetworkConfig := NetworkSelectionElements(networkConfig...)
	if podNetworkConfig == nil {
		return map[string]string{}
	}
	podNetworksConfig, err := json.Marshal(podNetworkConfig)
	if err != nil {
		return map[string]string{}
	}
	return map[string]string{
		nettypes.NetworkAttachmentAnnot: string(podNetworksConfig),
	}
}

func NetworkSelectionElements(networkConfig ...dynamicNetworkInfo) []nettypes.NetworkSelectionElement {
	var podNetworkConfig []nettypes.NetworkSelectionElement
	for i := range networkConfig {
		podNetworkConfig = append(
			podNetworkConfig,
			nettypes.NetworkSelectionElement{
				Name:             networkConfig[i].networkName,
				Namespace:        networkConfig[i].namespace,
				InterfaceRequest: networkConfig[i].ifaceName,
			},
		)
	}
	return podNetworkConfig
}

func namespacedName(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func ipWithMask(ip string, netmaskLen int) string {
	return fmt.Sprintf("%s/%d", ip, netmaskLen)
}

func lowerDeviceName() string {
	const (
		defaultLowerDeviceIfaceName = "eth0"
		lowerDeviceEnvVarKeyName    = "LOWER_DEVICE"
	)

	if lowerDeviceIfaceName, wasFound := os.LookupEnv(lowerDeviceEnvVarKeyName); wasFound {
		return lowerDeviceIfaceName
	}
	return defaultLowerDeviceIfaceName
}

func appLabel(appName string) string {
	return fmt.Sprintf("app=%s", appName)
}
