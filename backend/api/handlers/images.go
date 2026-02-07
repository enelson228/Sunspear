package handlers

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

func (h *ImageHandler) TagImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]

	var req struct {
		Repo string `json:"repo"`
		Tag  string `json:"tag"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Repo == "" {
		http.Error(w, "repo is required", http.StatusBadRequest)
		return
	}

	if req.Tag == "" {
		req.Tag = "latest"
	}

	newRef := fmt.Sprintf("%s:%s", req.Repo, req.Tag)

	if err := h.dockerService.TagImage(r.Context(), imageID, newRef); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "tagged", "tag": newRef})
}

func (h *ImageHandler) InspectImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]

	inspect, err := h.dockerService.InspectImage(r.Context(), imageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, inspect)
}

func (h *ImageHandler) GetImageHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]

	history, err := h.dockerService.GetImageHistory(r.Context(), imageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, history)
}

func (h *ImageHandler) PruneImages(w http.ResponseWriter, r *http.Request) {
	report, err := h.dockerService.PruneImages(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"deleted":        report.ImagesDeleted,
		"spaceReclaimed": report.SpaceReclaimed,
	})
}

func (h *ImageHandler) BuildImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dockerfile := r.FormValue("dockerfile")
	if dockerfile == "" {
		http.Error(w, "dockerfile content is required", http.StatusBadRequest)
		return
	}

	tagsStr := r.FormValue("tags")
	if tagsStr == "" {
		http.Error(w, "at least one tag is required", http.StatusBadRequest)
		return
	}

	tags := strings.Split(tagsStr, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	dockerfileBytes := []byte(dockerfile)
	if err := tw.WriteHeader(&tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfileBytes)),
		Mode: 0644,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := tw.Write(dockerfileBytes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tw.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buildResponse, err := h.dockerService.BuildImage(r.Context(), buf, tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer buildResponse.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, buildResponse.Body)
}
