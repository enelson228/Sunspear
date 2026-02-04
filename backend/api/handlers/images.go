package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"sunspear/services"

	"github.com/gorilla/mux"
)

type ImageHandler struct {
	dockerService *services.DockerService
}

func NewImageHandler(dockerService *services.DockerService) *ImageHandler {
	return &ImageHandler{dockerService: dockerService}
}

func (h *ImageHandler) ListImages(w http.ResponseWriter, r *http.Request) {
	images, err := h.dockerService.ListImages(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, images)
}

func (h *ImageHandler) PullImage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Image string `json:"image"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reader, err := h.dockerService.PullImage(r.Context(), req.Image)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, reader)
}

func (h *ImageHandler) RemoveImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]

	force := r.URL.Query().Get("force") == "true"

	response, err := h.dockerService.RemoveImage(r.Context(), imageID, force)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *ImageHandler) SearchImages(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")
	if term == "" {
		http.Error(w, "search term required", http.StatusBadRequest)
		return
	}

	results, err := h.dockerService.SearchImages(r.Context(), term)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, results)
}
