// mixed_mode_test.go contains E2E tests that validate NCP mixed-mode (T1 + VPC
// namespaces coexist) scope-gate behaviour.
//
// These tests are NOT part of the standard pre-check-in suite. They must be
// enabled explicitly with the -run-mixed-mode flag:
//
//	e2e=true go test -v ./test/e2e -run TestMixedMode_ \
//	    -run-mixed-mode \
//	    -remote.kubeconfig /root/.kube/config \
//	    -vc-user <user> -vc-password <password> \
//	    -test.timeout 30m
//
// Environment requirements:
//   - WCP Supervisor with SupervisorCapabilities CR accessible via kubectl.
//   - NCP (nsx-ujo) built from the m1-with-gate branch deployed on the Supervisor.
//   - NSX Manager reachable from the test runner.

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/vmware-tanzu/nsx-operator/pkg/nsx/services/common"
)

// mixedModeSkip skips the test when -run-mixed-mode is not set.
func mixedModeSkip(t *testing.T) {
	t.Helper()
	if !testOptions.runMixedMode {
		t.Skip("Skipping mixed-mode tests. Re-run with -run-mixed-mode to enable.")
	}
}

// setupMixedMode enables the supports_per_namespace_network_providers capability
// and creates unique namespaces for this test run. Each test gets its own T1 and VPC
// namespaces to avoid naming conflicts when tests run sequentially.
// It returns a cleanup function that should be deferred by the caller.
func setupMixedMode(t *testing.T) (func(), string, string) {
	t.Helper()

	// Generate unique namespace names for this test run
	// (avoid reusing the global constants which would cause "namespace already exists" errors)
	mmT1  := "e2e-mm-t1-" + getRandomString()
	mmVPC := "e2e-mm-vpc-" + getRandomString()

	t.Log("Enabling supports_per_namespace_network_providers on SupervisorCapabilities CR")
	require.NoError(t, hackSupervisorCapability("supports_per_namespace_network_providers", true),
		"failed to enable supervisor capability")

	// 2. Create the two test namespaces via VC API. When supports_per_namespace_network_providers
	//    is enabled, NCP automatically adds vpc_network_config annotation to ALL new namespaces.
	//    We then manipulate these annotations to simulate T1 vs VPC scope:
	//    - T1 namespace: remove vpc_network_config annotation
	//    - VPC namespace: keep vpc_network_config annotation (added by NCP automatically)
	t.Logf("Creating VC namespace for T1 simulation: %s", mmT1)
	if err := testData.createVCNamespace(mmT1); err != nil {
		t.Fatalf("Failed to create VC namespace %s: %v", mmT1, err)
	}
	t.Logf("Creating VC namespace for VPC simulation: %s", mmVPC)
	if err := testData.createVCNamespace(mmVPC); err != nil {
		t.Fatalf("Failed to create VC namespace %s: %v", mmVPC, err)
	}

	// 3. Remove vpc_network_config annotation from T1 namespace to make NCP classify it as T1 scope.
	//    VPC namespace already has vpc_network_config (added automatically by NCP), so it will be
	//    classified as VPC scope.
	t.Logf("Removing vpc_network_config annotation from %s to simulate T1 scope", mmT1)
	require.NoError(t,
		testData.patchNamespaceAnnotation(mmT1, common.AnnotationVPCNetworkConfig, ""),
		"failed to remove vpc_network_config annotation")
	
	// Verify the annotation was actually removed
	ns, err := testData.clientset.CoreV1().Namespaces().Get(context.TODO(), mmT1, metav1.GetOptions{})
	require.NoError(t, err, "failed to get namespace after annotation removal")
	if val, exists := ns.Annotations[common.AnnotationVPCNetworkConfig]; exists && val != "" {
		t.Logf("WARNING: vpc_network_config annotation still exists on T1 namespace after deletion: %v", val)
	} else {
		t.Logf("✓ vpc_network_config annotation successfully removed from %s", mmT1)
	}

	// 4. Wait for NCP to pick up the annotation change (refresh interval is 30s).
	//    After this wait, NCP should have updated its mixed-mode state:
	//    - mmT1: classified as T1 scope (no vpc_network_config)
	//    - mmVPC: classified as VPC scope (has vpc_network_config)
	t.Log("Waiting 35s for NCP mixed-mode refresh cycle to detect the annotation change")
	time.Sleep(35 * time.Second)

	return func() {
		t.Log("Cleaning up mixed-mode test namespaces and capability")
		CleanupVCNamespaces(mmT1, mmVPC)
		// Restore capability to disabled so it does not affect other tests.
		if err := hackSupervisorCapability("supports_per_namespace_network_providers", false); err != nil {
			t.Logf("Warning: failed to disable supervisor capability during cleanup: %v", err)
		}
	}, mmT1, mmVPC
}

// TestMixedMode_T1Controller_Normal verifies that T1 controllers process
// resources created in the T1 namespace correctly (baseline check).
func TestMixedMode_T1Controller_Normal(t *testing.T) {
	TrackTest(t)
	mixedModeSkip(t)

	cleanup, mmT1, mmVPC := setupMixedMode(t)
	defer cleanup()
	_ = mmVPC  // VPC namespace is created but not used in this test

	ns := mmT1
	podName := "mm-t1-pod"
	svcName := "mm-t1-svc"

	// Create a pod in the T1 namespace.
	t.Logf("Creating pod %s in T1 namespace %s", podName, ns)
	_, err := testData.createPod(ns, podName, "c1", "nginx", corev1.ProtocolTCP, 80)
	require.NoError(t, err, "failed to create pod in T1 namespace")
	defer func() {
		_ = testData.clientset.CoreV1().Pods(ns).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	}()

	// Create a service in the T1 namespace.
	t.Logf("Creating service %s in T1 namespace %s", svcName, ns)
	_, err = testData.createService(ns, svcName, 80, 80, corev1.ProtocolTCP, map[string]string{"app": "mm-t1"}, corev1.ServiceTypeClusterIP)
	require.NoError(t, err, "failed to create service in T1 namespace")
	defer func() {
		_ = testData.deleteService(ns, svcName)
	}()

	// Verify pod is synced to NSX inventory as ContainerApplicationInstance.
	RunSubtest(t, "PodInInventory", func(t *testing.T) {
		err := testData.waitForResourceExistOrNot(ns, "ContainerApplicationInstance", podName, true)
		assert.NoError(t, err, "T1 pod should be present in NSX inventory")
	})

	// Verify service is synced to NSX inventory as ContainerApplication.
	RunSubtest(t, "ServiceInInventory", func(t *testing.T) {
		err := testData.waitForResourceExistOrNot(ns, "ContainerApplication", svcName, true)
		assert.NoError(t, err, "T1 service should be present in NSX inventory")
	})
}

// TestMixedMode_T1Controller_IgnoreVPC verifies that T1 controllers do NOT
// process resources that belong to the VPC namespace.
func TestMixedMode_T1Controller_IgnoreVPC(t *testing.T) {
	TrackTest(t)
	mixedModeSkip(t)

	cleanup, mmT1, mmVPC := setupMixedMode(t)
	defer cleanup()
	_ = mmT1  // T1 namespace is created but not used in this test

	ns := mmVPC
	podName := "mm-vpc-pod"
	svcName := "mm-vpc-svc"

	// Create a pod in the VPC namespace.
	nsObj, _ := testData.clientset.CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{})
	t.Logf("VPC Namespace annotations: %v", nsObj.Annotations)
	t.Logf("Creating pod %s in VPC namespace %s", podName, ns)
	_, err := testData.createPod(ns, podName, "c1", "nginx", corev1.ProtocolTCP, 80)
	require.NoError(t, err, "failed to create pod in VPC namespace")
	defer func() {
		_ = testData.clientset.CoreV1().Pods(ns).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	}()

	// Create a service in the VPC namespace.
	t.Logf("Creating service %s in VPC namespace %s", svcName, ns)
	_, err = testData.createService(ns, svcName, 80, 80, corev1.ProtocolTCP, map[string]string{"app": "mm-vpc"}, corev1.ServiceTypeClusterIP)
	require.NoError(t, err, "failed to create service in VPC namespace")
	defer func() {
		_ = testData.deleteService(ns, svcName)
	}()

	// The T1 inventory controller should NOT sync VPC resources.
	// Wait the full default timeout; if nothing appears that is the desired outcome.
	RunSubtest(t, "PodNotInInventory", func(t *testing.T) {
		err := testData.waitForResourceExistOrNot(ns, "ContainerApplicationInstance", podName, false)
		assert.NoError(t, err, "VPC pod should NOT be present in NSX T1 inventory")
	})

	RunSubtest(t, "ServiceNotInInventory", func(t *testing.T) {
		err := testData.waitForResourceExistOrNot(ns, "ContainerApplication", svcName, false)
		assert.NoError(t, err, "VPC service should NOT be present in NSX T1 inventory")
	})
}

// TestMixedMode_VPCController_IgnoreT1 verifies that the VPCNamespaceController
// does NOT create VPC topology (Subnet) for the T1 namespace.
func TestMixedMode_VPCController_IgnoreT1(t *testing.T) {
	TrackTest(t)
	mixedModeSkip(t)

	cleanup, mmT1, mmVPC := setupMixedMode(t)
	defer cleanup()
	_ = mmVPC  // VPC namespace is created but not used in this test

	ns := mmT1

	// Allow enough time for VPCNamespaceController to process the namespace event.
	// If it were not gated, it would create a SubnetSet within defaultTimeout.
	RunSubtest(t, "NoSubnetSetForT1Namespace", func(t *testing.T) {
		// We expect NO SubnetSet to appear for the T1 namespace.
		// waitForResourceExistOrNot with shouldExist=false will succeed immediately
		// if nothing is found, or keep polling until timeout if something is wrongly created.
		err := wait.PollUntilContextTimeout(context.TODO(), 2*time.Second, defaultTimeout, false,
			func(ctx context.Context) (bool, error) {
				// Check that no SubnetSet tagged with this namespace exists.
				searchErr := testData.waitForResourceExist(ns, "SubnetSet", "display_name", fmt.Sprintf("%s-default", ns), false)
				if searchErr == nil {
					// confirmed absent
					return true, nil
				}
				return false, nil
			})
		assert.NoError(t, err, "VPC SubnetSet should NOT be created for T1 namespace %s", ns)
	})
}

// TestMixedMode_Inventory_Isolation verifies that resources created in the VPC
// namespace are NOT synced to the NSX Container Inventory (K8sInventoryController
// scope-gate fix).
func TestMixedMode_Inventory_Isolation(t *testing.T) {
	TrackTest(t)
	mixedModeSkip(t)

	cleanup, mmT1, mmVPC := setupMixedMode(t)
	defer cleanup()

	ns := mmVPC
	podName := "mm-inv-pod"
	svcName := "mm-inv-svc"

	// Create resources in the VPC namespace.
	t.Logf("Creating pod %s and service %s in VPC namespace %s", podName, svcName, ns)
	_, err := testData.createPod(ns, podName, "c1", "nginx", corev1.ProtocolTCP, 80)
	require.NoError(t, err, "failed to create pod")
	defer func() {
		_ = testData.clientset.CoreV1().Pods(ns).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	}()

	_, err = testData.createService(ns, svcName, 80, 80, corev1.ProtocolTCP, map[string]string{"app": "mm-inv"}, corev1.ServiceTypeClusterIP)
	require.NoError(t, err, "failed to create service")
	defer func() {
		_ = testData.deleteService(ns, svcName)
	}()

	// Confirm neither pod nor service appears in NSX Container Inventory.
	RunSubtest(t, "VPCPodNotInInventory", func(t *testing.T) {
		err := testData.waitForResourceExistOrNot(ns, "ContainerApplicationInstance", podName, false)
		assert.NoError(t, err, "VPC pod should NOT be synced to NSX Container Inventory")
	})

	RunSubtest(t, "VPCServiceNotInInventory", func(t *testing.T) {
		err := testData.waitForResourceExistOrNot(ns, "ContainerApplication", svcName, false)
		assert.NoError(t, err, "VPC service should NOT be synced to NSX Container Inventory")
	})

	// Also verify the T1 namespace IS still present in inventory (regression guard).
	RunSubtest(t, "T1NamespaceStillInInventory", func(t *testing.T) {
		err := testData.waitForResourceExistOrNot(mmT1, "ContainerProject", mmT1, true)
		assert.NoError(t, err, "T1 namespace should still be synced to NSX inventory")
	})
}
