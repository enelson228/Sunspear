package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sunspear/services"

	"github.com/gorilla/mux"
)

type ComposeHandler struct {
	composeService *services.ComposeService
}

func NewComposeHandler(composeService *services.ComposeService) *ComposeHandler {
	return &ComposeHandler{composeService: composeService}
}

func (h *ComposeHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.composeService.ListProjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, projects)
}

func (h *ComposeHandler) DeployProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		YAML        string `json:"yaml"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.YAML == "" {
		http.Error(w, "yaml is required", http.StatusBadRequest)
		return
	}

	project, err := h.composeService.Deploy(r.Context(), req.Name, req.Description, req.YAML)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, project)
}

func (h *ComposeHandler) ValidateYAML(w http.ResponseWriter, r *http.Request) {
	var req struct {
		YAML string `json:"yaml"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.YAML == "" {
		http.Error(w, "yaml is required", http.StatusBadRequest)
		return
	}

	composeFile, err := h.composeService.ParseYAML(req.YAML)
	if err != nil {
		http.Error(w, "Invalid YAML: "+err.Error(), http.StatusBadRequest)
		return
	}

	serviceNames := make([]string, 0, len(composeFile.Services))
	for name := range composeFile.Services {
		serviceNames = append(serviceNames, name)
	}

	response := map[string]interface{}{
		"valid":    true,
		"services": serviceNames,
		"version":  composeFile.Version,
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *ComposeHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.composeService.ListTemplates()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, templates)
}

func (h *ComposeHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	template, err := h.composeService.GetTemplate(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, template)
}

func (h *ComposeHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := h.composeService.GetProject(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, project)
}

func (h *ComposeHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	if err := h.composeService.DeleteProject(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *ComposeHandler) StartProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	if err := h.composeService.StartProject(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (h *ComposeHandler) StopProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	if err := h.composeService.StopProject(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (h *ComposeHandler) RestartProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	if err := h.composeService.RestartProject(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "restarted"})
}
