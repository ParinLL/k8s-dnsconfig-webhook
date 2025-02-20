package admission

import (
	"encoding/json"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// DNSConfigMutator handles DNS configuration mutations
type DNSConfigMutator struct {
	ndotsValue string
}

// NewDNSConfigMutator creates a new DNSConfigMutator
func NewDNSConfigMutator(ndotsValue string) *DNSConfigMutator {
	return &DNSConfigMutator{
		ndotsValue: ndotsValue,
	}
}

// Mutate modifies the pod's DNS configuration
func (m *DNSConfigMutator) Mutate(rawPod []byte) (*admissionv1.AdmissionResponse, error) {
	// Decode the pod
	var pod corev1.Pod
	if err := json.Unmarshal(rawPod, &pod); err != nil {
		klog.Errorf("Failed to unmarshal pod: %v", err)
		return AdmissionError(err)
	}

	// Create patch operations
	var patches []PatchOperation

	// Define the desired DNS configuration
	dnsConfig := &corev1.PodDNSConfig{
		Options: []corev1.PodDNSConfigOption{
			{
				Name:  "ndots",
				Value: &m.ndotsValue,
			},
		},
	}

	// Determine if we need to add or replace the DNS config
	if pod.Spec.DNSConfig == nil {
		patches = append(patches, PatchOperation{
			Op:    "add",
			Path:  "/spec/dnsConfig",
			Value: dnsConfig,
		})
	} else {
		patches = append(patches, PatchOperation{
			Op:    "replace",
			Path:  "/spec/dnsConfig",
			Value: dnsConfig,
		})
	}

	return AdmissionSuccess(patches)
}
