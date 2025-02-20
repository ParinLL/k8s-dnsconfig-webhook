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
	return &Handler{
		mutator: mutator,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Verify content type
	if r.Header.Get("Content-Type") != "application/json" {
		klog.Errorf("Invalid content type: %s", r.Header.Get("Content-Type"))
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

	// Process the admission request
	admissionResponse, err := h.mutator.Mutate(admissionReview.Request.Object.Raw)
	if err != nil {
		klog.Errorf("Mutation failed: %v", err)
		http.Error(w, "mutation failed", http.StatusInternalServerError)
		return
	}

	// Set the admission review response
	admissionReview.Response = admissionResponse
	admissionReview.Response.UID = admissionReview.Request.UID

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
