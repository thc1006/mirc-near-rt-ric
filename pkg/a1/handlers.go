package a1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hctsai1006/near-rt-ric/pkg/a1/enrichment"
	a1_generated "github.com/hctsai1006/near-rt-ric/pkg/a1/generated"
	"github.com/hctsai1006/near-rt-ric/pkg/a1/mlmodel"
	"github.com/hctsai1006/near-rt-ric/pkg/a1/policy"
)

// Implementations of a1_api.ServerInterface
func (s *A1Server) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var newPolicy a1_generated.Policy
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.logger.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &newPolicy)
	if err != nil {
		s.logger.Errorf("Failed to unmarshal policy: %v", err)
		http.Error(w, "Invalid policy data", http.StatusBadRequest)
		return
	}

	// Set creation and update timestamps
	now := time.Now()
	newPolicy.CreatedAt = &now
	newPolicy.UpdatedAt = &now
	status := string(policy.PolicyStatusDraft)
	newPolicy.Status = &status // Initial status

	if newPolicy.Id == nil || *newPolicy.Id == "" {
		// Generate a UUID for the policy if not provided
		id := fmt.Sprintf("policy-%d", time.Now().UnixNano())
		newPolicy.Id = &id
	}

	p := policy.Policy{
		ID:        *newPolicy.Id,
		Name:      newPolicy.Name,
		Version:   newPolicy.Version,
		Content:   newPolicy.Content,
		Status:    policy.PolicyStatus(*newPolicy.Status),
		CreatedAt: newPolicy.CreatedAt.Format(time.RFC3339),
		UpdatedAt: newPolicy.UpdatedAt.Format(time.RFC3339),
	}

	err = s.policyMgr.CreatePolicy(r.Context(), &p)
	if err != nil {
		s.logger.Errorf("Failed to create policy: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create policy: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPolicy)
}

func (s *A1Server) GetAllPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := s.policyMgr.GetAllPolicies(r.Context())
	if err != nil {
		s.logger.Errorf("Failed to get all policies: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get all policies: %v", err), http.StatusInternalServerError)
		return
	}

	var apiPolicies []a1_generated.Policy
	for _, p := range policies {
		status := string(p.Status)
		createdAt, _ := time.Parse(time.RFC3339, p.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, p.UpdatedAt)
		apiPolicies = append(apiPolicies, a1_generated.Policy{
			Id:        &p.ID,
			Name:      p.Name,
			Version:   p.Version,
			Content:   p.Content,
			Status:    &status,
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiPolicies)
}

func (s *A1Server) GetPolicyById(w http.ResponseWriter, r *http.Request, id string) {
	p, err := s.policyMgr.GetPolicy(r.Context(), id)
	if err != nil {
		s.logger.Errorf("Failed to get policy %s: %v", id, err)
		http.Error(w, fmt.Sprintf("Failed to get policy: %v", err), http.StatusNotFound)
		return
	}

	status := string(p.Status)
	createdAt, _ := time.Parse(time.RFC3339, p.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, p.UpdatedAt)
	apiPolicy := a1_generated.Policy{
		Id:        &p.ID,
		Name:      p.Name,
		Version:   p.Version,
		Content:   p.Content,
		Status:    &status,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiPolicy)
}

func (s *A1Server) UpdatePolicyById(w http.ResponseWriter, r *http.Request, id string) {
	var updatedPolicy a1_generated.Policy
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.logger.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &updatedPolicy)
	if err != nil {
		s.logger.Errorf("Failed to unmarshal policy: %v", err)
		http.Error(w, "Invalid policy data", http.StatusBadRequest)
		return
	}

	if updatedPolicy.Id == nil || *updatedPolicy.Id != id {
		http.Error(w, "Policy ID in path and body do not match", http.StatusBadRequest)
		return
	}

	now := time.Now()
	updatedPolicy.UpdatedAt = &now

	p := policy.Policy{
		ID:        *updatedPolicy.Id,
		Name:      updatedPolicy.Name,
		Version:   updatedPolicy.Version,
		Content:   updatedPolicy.Content,
		Status:    policy.PolicyStatus(*updatedPolicy.Status),
		CreatedAt: updatedPolicy.CreatedAt.Format(time.RFC3339),
		UpdatedAt: updatedPolicy.UpdatedAt.Format(time.RFC3339),
	}

	err = s.policyMgr.UpdatePolicy(r.Context(), &p)
	if err != nil {
		s.logger.Errorf("Failed to update policy %s: %v", id, err)
		http.Error(w, fmt.Sprintf("Failed to update policy: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedPolicy)
}

func (s *A1Server) DeletePolicyById(w http.ResponseWriter, r *http.Request, id string) {
	err := s.policyMgr.DeletePolicy(r.Context(), id)
	if err != nil {
		s.logger.Errorf("Failed to delete policy %s: %v", id, err)
		http.Error(w, fmt.Sprintf("Failed to delete policy: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *A1Server) GetPolicyTypes(w http.ResponseWriter, r *http.Request) {
	// For now, return a hardcoded list of policy types
	policyTypes := []string{
		"ORAN_E2_POLICY_V1",
		"ORAN_QOS_POLICY_V1",
		"ORAN_SLICE_POLICY_V1",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policyTypes)
}

func (s *A1Server) DeployMLModel(w http.ResponseWriter, r *http.Request) {
	var newModel a1_generated.MLModel
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.logger.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &newModel)
	if err != nil {
		s.logger.Errorf("Failed to unmarshal ML model: %v", err)
		http.Error(w, "Invalid ML model data", http.StatusBadRequest)
		return
	}

	// Set creation and update timestamps
	now := time.Now()
	newModel.CreatedAt = &now
	newModel.UpdatedAt = &now
	status := string(mlmodel.MLModelStatusNew)
	newModel.Status = &status // Initial status

	if newModel.Id == nil || *newModel.Id == "" {
		// Generate a UUID for the model if not provided
		id := fmt.Sprintf("model-%d", time.Now().UnixNano())
		newModel.Id = &id
	}

	m := mlmodel.MLModel{
		ID:        *newModel.Id,
		Name:      newModel.Name,
		Version:   newModel.Version,
		FilePath:  newModel.FilePath,
		Status:    mlmodel.MLModelStatus(*newModel.Status),
		CreatedAt: newModel.CreatedAt.Format(time.RFC3339),
		UpdatedAt: newModel.UpdatedAt.Format(time.RFC3339),
	}

	err = s.modelMgr.UploadModel(r.Context(), &m)
	if err != nil {
		s.logger.Errorf("Failed to upload ML model: %v", err)
		http.Error(w, fmt.Sprintf("Failed to upload ML model: %v", err), http.StatusInternalServerError)
		return
	}

	err = s.modelMgr.DeployModel(r.Context(), *newModel.Id)
	if err != nil {
		s.logger.Errorf("Failed to deploy ML model %s: %v", *newModel.Id, err)
		http.Error(w, fmt.Sprintf("Failed to deploy ML model: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newModel)
}

func (s *A1Server) GetAllMLModels(w http.ResponseWriter, r *http.Request) {
	models, err := s.modelMgr.GetAllModels(r.Context())
	if err != nil {
		s.logger.Errorf("Failed to get all ML models: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get all ML models: %v", err), http.StatusInternalServerError)
		return
	}

	var apiModels []a1_generated.MLModel
	for _, m := range models {
		status := string(m.Status)
		createdAt, _ := time.Parse(time.RFC3339, m.CreatedAt)
		updatedAt, _ := time.Parse(time.RFC3339, m.UpdatedAt)
		apiModels = append(apiModels, a1_generated.MLModel{
			Id:        &m.ID,
			Name:      m.Name,
			Version:   m.Version,
			FilePath:  m.FilePath,
			Status:    &status,
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiModels)
}

func (s *A1Server) GetMLModelById(w http.ResponseWriter, r *http.Request, id string) {
	m, err := s.modelMgr.GetModel(r.Context(), id)
	if err != nil {
		s.logger.Errorf("Failed to get ML model %s: %v", id, err)
		http.Error(w, fmt.Sprintf("Failed to get ML model: %v", err), http.StatusNotFound)
		return
	}

	status := string(m.Status)
	createdAt, _ := time.Parse(time.RFC3339, m.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, m.UpdatedAt)
	apiModel := a1_generated.MLModel{
		Id:        &m.ID,
		Name:      m.Name,
		Version:   m.Version,
		FilePath:  m.FilePath,
		Status:    &status,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiModel)
}

func (s *A1Server) RollbackMLModel(w http.ResponseWriter, r *http.Request, id string) {
	err := s.modelMgr.RollbackModel(r.Context(), id)
	if err != nil {
		s.logger.Errorf("Failed to rollback ML model %s: %v", id, err)
		http.Error(w, fmt.Sprintf("Failed to rollback ML model: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("ML Model %s rolled back successfully", id)))
}

func (s *A1Server) UpdateEnrichmentInfo(w http.ResponseWriter, r *http.Request) {
	var newInfo a1_generated.EnrichmentInfo
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.logger.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &newInfo)
	if err != nil {
		s.logger.Errorf("Failed to unmarshal enrichment info: %v", err)
		http.Error(w, "Invalid enrichment info data", http.StatusBadRequest)
		return
	}

	// Set timestamp and default TTL if not provided
	now := time.Now()
	newInfo.Timestamp = &now
	if newInfo.Ttl == nil || *newInfo.Ttl == 0 {
		ttl := int64(5 * time.Minute) // Default TTL
		newInfo.Ttl = &ttl
	}

	if newInfo.Id == nil || *newInfo.Id == "" {
		// Generate a UUID for the info if not provided
		id := fmt.Sprintf("enrichment-%d", time.Now().UnixNano())
		newInfo.Id = &id
	}

	ei := enrichment.EnrichmentInfo{
		ID:        *newInfo.Id,
		Name:      newInfo.Name,
		Source:    newInfo.Source,
		Content:   newInfo.Content,
		Timestamp: *newInfo.Timestamp,
		TTL:       time.Duration(*newInfo.Ttl),
	}

	err = s.enrichMgr.UpdateInfo(r.Context(), &ei)
	if err != nil {
		s.logger.Errorf("Failed to update enrichment info: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update enrichment info: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newInfo)
}

func (s *A1Server) GetAllEnrichmentInfo(w http.ResponseWriter, r *http.Request) {
	infoList, err := s.enrichMgr.GetAllInfo(r.Context())
	if err != nil {
		s.logger.Errorf("Failed to get all enrichment info: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get all enrichment info: %v", err), http.StatusInternalServerError)
		return
	}

	var apiInfoList []a1_generated.EnrichmentInfo
	for _, ei := range infoList {
		ttl := int64(ei.TTL)
		apiInfoList = append(apiInfoList, a1_generated.EnrichmentInfo{
			Id:        &ei.ID,
			Name:      ei.Name,
			Source:    ei.Source,
			Content:   ei.Content,
			Timestamp: &ei.Timestamp,
			Ttl:       &ttl,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiInfoList)
}

func (s *A1Server) GetEnrichmentInfoById(w http.ResponseWriter, r *http.Request, id string) {
	info, err := s.enrichMgr.GetInfo(r.Context(), id)
	if err != nil {
		s.logger.Errorf("Failed to get enrichment info %s: %v", id, err)
		http.Error(w, fmt.Sprintf("Failed to get enrichment info: %v", err), http.StatusNotFound)
		return
	}

	ttl := int64(info.TTL)
	apiInfo := a1_generated.EnrichmentInfo{
		Id:        &info.ID,
		Name:      info.Name,
		Source:    info.Source,
		Content:   info.Content,
		Timestamp: &info.Timestamp,
		Ttl:       &ttl,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiInfo)
}
