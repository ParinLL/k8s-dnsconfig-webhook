package admission

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// DNSConfig represents the structure of our DNS configuration
type DNSConfig struct {
	DNSConfig corev1.PodDNSConfig `yaml:"dnsConfig"`
}

// DNSConfigMutator handles DNS configuration mutations
type DNSConfigMutator struct {
	config *DNSConfig
}

// NewDNSConfigMutator creates a new DNSConfigMutator
func NewDNSConfigMutator(configData []byte) (*DNSConfigMutator, error) {
	config := &DNSConfig{}
	if err := yaml.Unmarshal(configData, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DNS config: %v", err)
	}

	return &DNSConfigMutator{
		config: config,
	}, nil
}

// Mutate modifies the pod's DNS configuration
func (m *DNSConfigMutator) Mutate(ar *admissionv1.AdmissionRequest) (*admissionv1.AdmissionResponse, error) {
	// Parse the Pod object
	var pod corev1.Pod
	if err := json.Unmarshal(ar.Object.Raw, &pod); err != nil {
		klog.Errorf("Failed to unmarshal pod: %v", err)
		return AdmissionError(fmt.Errorf("could not unmarshal pod object: %v", err))
	}

	podNamespace := ar.Namespace
	podName := pod.Name
	if podName == "" {
		podName = "<generating>"
	}

	klog.V(2).Infof("Processing mutation for pod %s/%s", podNamespace, podName)

	// Create patch operations
	var patches []PatchOperation

	// Log current DNS config
	if pod.Spec.DNSConfig != nil {
		currentConfig, _ := json.MarshalIndent(pod.Spec.DNSConfig, "", "  ")
		klog.V(2).Infof("Current DNS config for pod %s/%s: %s", podNamespace, podName, string(currentConfig))
	} else {
		klog.V(2).Infof("No existing DNS config for pod %s/%s", podNamespace, podName)
	}

	// Determine if we need to add or replace the DNS config
	if pod.Spec.DNSConfig == nil {
		patches = append(patches, PatchOperation{
			Op:    "add",
			Path:  "/spec/dnsConfig",
			Value: m.config.DNSConfig,
		})
		klog.V(2).Infof("Adding new DNS config to pod %s/%s", podNamespace, podName)
	} else {
		patches = append(patches, PatchOperation{
			Op:    "replace",
			Path:  "/spec/dnsConfig",
			Value: m.config.DNSConfig,
		})
		klog.V(2).Infof("Replacing DNS config for pod %s/%s", podNamespace, podName)
	}

	// Log the patches
	if len(patches) > 0 {
		patchesJSON, _ := json.MarshalIndent(patches, "", "  ")
		klog.V(2).Infof("Applying patches to pod %s/%s: %s", podNamespace, podName, string(patchesJSON))
	}

	return AdmissionSuccess(patches)
}

// UpdateConfig updates the DNS configuration
func (m *DNSConfigMutator) UpdateConfig(configData []byte) error {
	config := &DNSConfig{}
	if err := yaml.Unmarshal(configData, config); err != nil {
		return fmt.Errorf("failed to unmarshal DNS config: %v", err)
	}
	m.config = config
	return nil
}
