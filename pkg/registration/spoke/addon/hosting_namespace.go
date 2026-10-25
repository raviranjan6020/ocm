package addon

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	addonv1beta1 "open-cluster-management.io/api/addon/v1beta1"
)

// Prototype for enhancement 191 (hosted addon runtime namespace). The naming is a copy of
// addon-framework utils.HostingNamespace; a real change would use the vendored helper.

// hostingAddonAnnotationKey records which ManagedClusterAddOn (<cluster>/<addon>) owns a hosting namespace.
const hostingAddonAnnotationKey = "addon.open-cluster-management.io/hosting-addon"

func hostingNamespaceName(clusterName, addonName string) string {
	h := sha256.New()
	h.Write([]byte("open-cluster-management.io/addon-instance/v1\x00"))
	for _, v := range []string{clusterName, addonName} {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(v)))
		h.Write(length[:])
		h.Write([]byte(v))
	}
	return "open-cluster-management-" + addonToken(addonName) + "-" + hex.EncodeToString(h.Sum(nil))[:20]
}

func addonToken(addonName string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(addonName) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	token := strings.Trim(b.String(), "-")
	if len(token) > 18 {
		token = strings.TrimRight(token[:18], "-")
	}
	if token == "" {
		token = "addon"
	}
	return token
}

// addonNamespace returns where the addon's registration Secret and Lease live. For an addon running outside the
// managed cluster, it is the per-instance hosting namespace when that namespace exists on the management cluster
// and its owner annotation names this addon instance. Otherwise it is the existing installation namespace.
func addonNamespace(ctx context.Context, namespaces corev1client.NamespacesGetter, clusterName string,
	addOn *addonv1beta1.ManagedClusterAddOn) string {
	if namespaces != nil && isAddonRunningOutsideManagedCluster(addOn) {
		name := hostingNamespaceName(clusterName, addOn.Name)
		ns, err := namespaces.Namespaces().Get(ctx, name, metav1.GetOptions{})
		if err == nil && ns.Annotations[hostingAddonAnnotationKey] == clusterName+"/"+addOn.Name {
			return name
		}
	}
	return getAddOnInstallationNamespace(addOn)
}
