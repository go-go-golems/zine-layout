package serve

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/zine-layout/pkg/repo"
)

// Project layout sequence routes ---------------------------------------------

func (s *Server) handleProjectLayoutSequences(w http.ResponseWriter, r *http.Request, projectID string) {
	switch r.Method {
	case http.MethodGet:
		if _, err := s.repos.Projects.Get(projectID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		sequences, err := s.repos.LayoutSequences.ListByProject(projectID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := make([]layoutSequenceResponse, 0, len(sequences))
		for _, seq := range sequences {
			resp = append(resp, layoutSequenceToResponse(seq))
		}
		writeJSON(w, http.StatusOK, map[string]any{"layout_sequences": resp})
	case http.MethodPost:
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" {
			name = "Untitled Layout Sequence"
		}
		sequence := &repo.LayoutSequence{
			ProjectID:   projectID,
			Name:        name,
			Description: strings.TrimSpace(req.Description),
		}
		if err := s.repos.LayoutSequences.Create(sequence); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"layout_sequence": layoutSequenceToResponse(sequence)})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Layout sequence resource ----------------------------------------------------

func (s *Server) handleLayoutSequenceRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/layout-sequences/")
	if path == "" || path == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	sequenceID := parts[0]
	if sequenceID == "" {
		http.NotFound(w, r)
		return
	}
	if len(parts) > 1 {
		if parts[1] != "items" {
			http.NotFound(w, r)
			return
		}
		if len(parts) == 2 {
			s.handleLayoutSequenceItems(w, r, sequenceID)
			return
		}
		position, err := strconv.Atoi(parts[2])
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid item position")
			return
		}
		s.handleLayoutSequenceItemDeletion(w, r, sequenceID, position)
		return
	}

	switch r.Method {
	case http.MethodGet:
		sequence, err := s.repos.LayoutSequences.Get(sequenceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items, err := s.repos.LayoutSequences.ListItems(sequenceID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"layout_sequence": layoutSequenceToResponse(sequence),
			"items":           layoutSequenceItemsToResponse(items),
		})
	case http.MethodPatch, http.MethodPut:
		sequence, err := s.repos.LayoutSequences.Get(sequenceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		var req struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				name = "Untitled Layout Sequence"
			}
			sequence.Name = name
		}
		if req.Description != nil {
			sequence.Description = strings.TrimSpace(*req.Description)
		}
		sequence.UpdatedAt = time.Now().UTC()
		if err := s.repos.LayoutSequences.Update(sequence); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"layout_sequence": layoutSequenceToResponse(sequence)})
	case http.MethodDelete:
		if err := s.repos.LayoutSequences.Delete(sequenceID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, PATCH, PUT, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Layout sequence items -------------------------------------------------------

func (s *Server) handleLayoutSequenceItems(w http.ResponseWriter, r *http.Request, sequenceID string) {
	sequence, err := s.repos.LayoutSequences.Get(sequenceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := s.repos.LayoutSequences.ListItems(sequenceID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": layoutSequenceItemsToResponse(items)})
	case http.MethodPost:
		var req struct {
			LaidOutImageID string `json:"laid_out_image_id"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if strings.TrimSpace(req.LaidOutImageID) == "" {
			respondError(w, http.StatusBadRequest, "laid_out_image_id is required")
			return
		}
		image, err := s.repos.LaidOutImages.Get(req.LaidOutImageID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondError(w, http.StatusBadRequest, "laid_out_image not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if image.ProjectID != sequence.ProjectID {
			respondError(w, http.StatusBadRequest, "laid_out_image does not belong to sequence project")
			return
		}
		item := &repo.LayoutSequenceItem{
			SequenceID:     sequenceID,
			LaidOutImageID: req.LaidOutImageID,
		}
		if err := s.repos.LayoutSequences.AddItem(item); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items, err := s.repos.LayoutSequences.ListItems(sequenceID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"items": layoutSequenceItemsToResponse(items)})
	case http.MethodPut:
		var req struct {
			Items []struct {
				LaidOutImageID string `json:"laid_out_image_id"`
			} `json:"items"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		replacements := make([]*repo.LayoutSequenceItem, 0, len(req.Items))
		for idx, item := range req.Items {
			if strings.TrimSpace(item.LaidOutImageID) == "" {
				respondError(w, http.StatusBadRequest, fmt.Sprintf("laid_out_image_id required for item %d", idx))
				return
			}
			image, err := s.repos.LaidOutImages.Get(item.LaidOutImageID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					respondError(w, http.StatusBadRequest, fmt.Sprintf("laid_out_image %s not found", item.LaidOutImageID))
					return
				}
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if image.ProjectID != sequence.ProjectID {
				respondError(w, http.StatusBadRequest, "laid_out_image does not belong to sequence project")
				return
			}
			replacements = append(replacements, &repo.LayoutSequenceItem{
				SequenceID:     sequenceID,
				LaidOutImageID: item.LaidOutImageID,
			})
		}
		if err := s.repos.LayoutSequences.ReplaceItems(sequenceID, replacements); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items, err := s.repos.LayoutSequences.ListItems(sequenceID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": layoutSequenceItemsToResponse(items)})
	default:
		w.Header().Set("Allow", "GET, POST, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleLayoutSequenceItemDeletion(w http.ResponseWriter, r *http.Request, sequenceID string, position int) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.repos.LayoutSequences.DeleteItem(sequenceID, position); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
