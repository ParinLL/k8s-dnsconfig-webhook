package admission

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/klog/v2"
)

// Handler handles admission webhook requests
type Handler struct {
	mutator *DNSConfigMutator
}

// NewHandler creates a new admission Handler
func NewHandler(mutator *DNSConfigMutator) *Handler {
	klog.V(2).Info("Creating new admission handler")
	return &Handler{
		mutator: mutator,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	klog.V(2).Info("=== Starting admission request handling ===")
	defer klog.V(2).Info("=== Completed admission request handling ===")

	// Verify content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("Invalid content type: %s", contentType)
		http.Error(w, "invalid Content-Type, want application/json", http.StatusBadRequest)
		return
	}

	// Read the body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		klog.Errorf("Failed to read request body: %v", err)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse the AdmissionReview
	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		klog.Errorf("Failed to unmarshal request: %v", err)
		http.Error(w, "failed to unmarshal request", http.StatusBadRequest)
		return
	}

	// Log request details
	klog.V(2).Infof("Admission request details:")
	klog.V(2).Infof("  UID: %s", admissionReview.Request.UID)
	klog.V(2).Infof("  Kind: %v", admissionReview.Request.Kind)
	klog.V(2).Infof("  Resource: %v", admissionReview.Request.Resource)
	klog.V(2).Infof("  Name: %s", admissionReview.Request.Name)
	klog.V(2).Infof("  Namespace: %s", admissionReview.Request.Namespace)
	klog.V(2).Infof("  Operation: %s", admissionReview.Request.Operation)
	klog.V(2).Infof("  UserInfo: %v", admissionReview.Request.UserInfo)

	// Process the admission request
	klog.V(2).Info("Processing mutation request")
	admissionResponse, err := h.mutator.Mutate(admissionReview.Request)
	if err != nil {
		klog.Errorf("Mutation failed: %v", err)
		http.Error(w, "mutation failed", http.StatusInternalServerError)
		return
	}

	// Set the admission review response
	admissionReview.Response = admissionResponse
	admissionReview.Response.UID = admissionReview.Request.UID

	// Log response details
	if admissionResponse.Allowed {
		if len(admissionResponse.Patch) > 0 {
			klog.V(2).Infof("Mutation applied successfully for %s/%s",
				admissionReview.Request.Namespace,
				admissionReview.Request.Name)
		} else {
			klog.V(2).Infof("No mutation needed for %s/%s",
				admissionReview.Request.Namespace,
				admissionReview.Request.Name)
		}
	} else {
		klog.V(2).Infof("Admission denied for %s/%s: %s",
			admissionReview.Request.Namespace,
			admissionReview.Request.Name,
			admissionResponse.Result.Message)
	}

	// Send response
	resp, err := json.Marshal(admissionReview)
	if err != nil {
		klog.Errorf("Failed to marshal response: %v", err)
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}
