package addon

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"

	addonv1beta1 "open-cluster-management.io/api/addon/v1beta1"
)

func TestAddonNamespace(t *testing.T) {
	const clusterName, addonName = "cluster-a", "helloworldhosted"
	hostingNS := hostingNamespaceName(clusterName, addonName)
	namespace := func(owner string) *corev1.Namespace {
		return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
			Name: hostingNS, Annotations: map[string]string{hostingAddonAnnotationKey: owner}}}
	}

	cases := []struct {
		name     string
		hosted   bool
		existing []runtime.Object
		want     string
	}{
		{name: "hosted, namespace owned by this addon", hosted: true,
			existing: []runtime.Object{namespace("cluster-a/helloworldhosted")}, want: hostingNS},
		{name: "hosted, namespace owned by another addon", hosted: true,
			existing: []runtime.Object{namespace("cluster-b/helloworldhosted")}, want: "open-cluster-management-agent-addon"},
		{name: "hosted, no hosting namespace", hosted: true, want: "open-cluster-management-agent-addon"},
		{name: "default mode", hosted: false,
			existing: []runtime.Object{namespace("cluster-a/helloworldhosted")}, want: "open-cluster-management-agent-addon"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			addOn := &addonv1beta1.ManagedClusterAddOn{ObjectMeta: metav1.ObjectMeta{
				Name: addonName, Namespace: clusterName, Annotations: map[string]string{}}}
			if c.hosted {
				addOn.Annotations[addonv1beta1.HostingClusterNameAnnotationKey] = "hosting"
			}
			client := kubefake.NewClientset(c.existing...)
			if got := addonNamespace(context.TODO(), client.CoreV1(), clusterName, addOn); got != c.want {
				t.Errorf("addonNamespace() = %q, want %q", got, c.want)
			}
			if !c.hosted && len(client.Actions()) != 0 {
				t.Errorf("expected no API calls for a default-mode addon, got %d", len(client.Actions()))
			}
		})
	}
}
