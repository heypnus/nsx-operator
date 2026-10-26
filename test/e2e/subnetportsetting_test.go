package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	restclient "k8s.io/client-go/rest"

	"github.com/vmware-tanzu/nsx-operator/pkg/apis/vpc/v1alpha1"
	"github.com/vmware-tanzu/nsx-operator/pkg/client/clientset/versioned"
	"github.com/vmware-tanzu/nsx-operator/pkg/nsx/services/common"
)

// TestSubnetPortSettingSync verifies synchronization of SubnetPortSetting CRs for shared Subnets
func TestSubnetPortSettingSync(t *testing.T) {
	TrackTest(t)

	// 1. Use pre-created VC namespaces (already recognized as VPC namespaces)
	ns1 := NsSubnetPrecreated1
	ns2 := NsSubnetPrecreated2
	associatedResource := "project-quality:kube-system_mo26n:vm-default-5dd6f5d2_mo26n"
	psID := "e2e-port-setting-1"

	// 2. Create PortSetting on NSX via PortSettingsClient
	orgID := "default"
	projectID := "project-quality"
	vpcID := "kube-system_mo26n"
	subnetID := "vm-default-5dd6f5d2_mo26n"

	nsxPS := model.VpcSubnetPortSetting{
		Id:          common.String(psID),
		DisplayName: common.String(psID),
	}
	err := testData.nsxClient.PortSettingsClient.Patch(orgID, projectID, vpcID, subnetID, psID, nsxPS)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = testData.nsxClient.PortSettingsClient.Delete(orgID, projectID, vpcID, subnetID, psID)
	})

	// Wait 2 seconds to ensure namespace informers are updated
	time.Sleep(2 * time.Second)

	// 3. Create impersonated CRD client as NSX Operator service account
	impersonatedCfg := restclient.CopyConfig(testData.kubeConfig)
	impersonatedCfg.Impersonate = restclient.ImpersonationConfig{
		UserName: "system:serviceaccount:vmware-system-nsx:ncp-svc-account",
	}
	opClientset, err := versioned.NewForConfig(impersonatedCfg)
	require.NoError(t, err)

	// Create shared Subnet CRs in ns1 and ns2
	subnetCR1 := &v1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shared-subnet-e2e-1",
			Namespace: ns1,
			Annotations: map[string]string{
				common.AnnotationAssociatedResource: associatedResource,
			},
		},
		Spec: v1alpha1.SubnetSpec{},
	}
	_, err = opClientset.CrdV1alpha1().Subnets(ns1).Create(context.TODO(), subnetCR1, metav1.CreateOptions{})
	require.NoError(t, err)

	subnetCR2 := &v1alpha1.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shared-subnet-e2e-2",
			Namespace: ns2,
			Annotations: map[string]string{
				common.AnnotationAssociatedResource: associatedResource,
			},
		},
		Spec: v1alpha1.SubnetSpec{},
	}
	_, err = opClientset.CrdV1alpha1().Subnets(ns2).Create(context.TODO(), subnetCR2, metav1.CreateOptions{})
	require.NoError(t, err)

	// Wait for SubnetCR1 and SubnetCR2 to be processed and Ready
	err = wait.PollUntilContextTimeout(context.TODO(), 2*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		s1, err1 := testData.crdClientset.CrdV1alpha1().Subnets(ns1).Get(ctx, subnetCR1.Name, metav1.GetOptions{})
		s2, err2 := testData.crdClientset.CrdV1alpha1().Subnets(ns2).Get(ctx, subnetCR2.Name, metav1.GetOptions{})
		if err1 != nil || err2 != nil {
			return false, nil
		}
		s1Ready, s2Ready := false, false
		for _, cond := range s1.Status.Conditions {
			if cond.Type == v1alpha1.Ready && cond.Status == corev1.ConditionTrue {
				s1Ready = true
			}
		}
		for _, cond := range s2.Status.Conditions {
			if cond.Type == v1alpha1.Ready && cond.Status == corev1.ConditionTrue {
				s2Ready = true
			}
		}
		t.Logf("Subnet readiness: ns1=%v, ns2=%v", s1Ready, s2Ready)
		return s1Ready && s2Ready, nil
	})
	require.NoError(t, err, "Shared subnets should become ready")

	// 4. Verify SubnetPortSetting CR created and Ready=True in ns1
	err = wait.PollUntilContextTimeout(context.TODO(), 2*time.Second, 90*time.Second, true, func(ctx context.Context) (bool, error) {
		list, listErr := testData.crdClientset.CrdV1alpha1().SubnetPortSettings(ns1).List(ctx, metav1.ListOptions{})
		if listErr != nil {
			t.Logf("List ns1 SubnetPortSettings error: %v", listErr)
			return false, nil
		}
		if len(list.Items) == 0 {
			t.Logf("List ns1 SubnetPortSettings empty, waiting...")
			return false, nil
		}
		for _, item := range list.Items {
			t.Logf("Found ns1 SubnetPortSetting %s, subnetName=%s, conditions=%+v", item.Name, item.Spec.SubnetName, item.Status.Conditions)
			if item.Spec.SubnetName == subnetCR1.Name {
				for _, cond := range item.Status.Conditions {
					if cond.Type == v1alpha1.Ready && cond.Status == corev1.ConditionTrue && cond.Reason == "PortSettingRealized" {
						return true, nil
					}
				}
			}
		}
		return false, nil
	})
	require.NoError(t, err, "SubnetPortSetting should be synchronized and Ready=True in ns1")

	// 5. Verify SubnetPortSetting CR created and Ready=True in ns2
	err = wait.PollUntilContextTimeout(context.TODO(), 2*time.Second, 90*time.Second, true, func(ctx context.Context) (bool, error) {
		list, listErr := testData.crdClientset.CrdV1alpha1().SubnetPortSettings(ns2).List(ctx, metav1.ListOptions{})
		if listErr != nil {
			t.Logf("List ns2 SubnetPortSettings error: %v", listErr)
			return false, nil
		}
		if len(list.Items) == 0 {
			t.Logf("List ns2 SubnetPortSettings empty, waiting...")
			return false, nil
		}
		for _, item := range list.Items {
			t.Logf("Found ns2 SubnetPortSetting %s, subnetName=%s, conditions=%+v", item.Name, item.Spec.SubnetName, item.Status.Conditions)
			if item.Spec.SubnetName == subnetCR2.Name {
				for _, cond := range item.Status.Conditions {
					if cond.Type == v1alpha1.Ready && cond.Status == corev1.ConditionTrue && cond.Reason == "PortSettingRealized" {
						return true, nil
					}
				}
			}
		}
		return false, nil
	})
	require.NoError(t, err, "SubnetPortSetting should be synchronized and Ready=True in ns2")

	// 6. Delete Subnet CR in ns1 and verify cascade deletion of SubnetPortSetting CR
	err = opClientset.CrdV1alpha1().Subnets(ns1).Delete(context.TODO(), subnetCR1.Name, metav1.DeleteOptions{})
	require.NoError(t, err)

	err = wait.PollUntilContextTimeout(context.TODO(), 2*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		list, listErr := testData.crdClientset.CrdV1alpha1().SubnetPortSettings(ns1).List(ctx, metav1.ListOptions{})
		return listErr == nil && len(list.Items) == 0, nil
	})
	require.NoError(t, err, "SubnetPortSetting in ns1 should be cascade deleted with parent Subnet")

	// 7. Verify SubnetPortSetting in ns2 is unaffected
	list2, err := testData.crdClientset.CrdV1alpha1().SubnetPortSettings(ns2).List(context.TODO(), metav1.ListOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, list2.Items)
	require.Equal(t, subnetCR2.Name, list2.Items[0].Spec.SubnetName)

	// Clean up ns2 subnet
	_ = opClientset.CrdV1alpha1().Subnets(ns2).Delete(context.TODO(), subnetCR2.Name, metav1.DeleteOptions{})
	_ = opClientset.CrdV1alpha1().Subnets(ns1).Delete(context.TODO(), subnetCR1.Name, metav1.DeleteOptions{})
}
