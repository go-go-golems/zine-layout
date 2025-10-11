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

// Project image sequence routes ----------------------------------------------

func (s *Server) handleProjectSequences(w http.ResponseWriter, r *http.Request, projectID string) {
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
		sequences, err := s.repos.ImageSequences.ListByProject(projectID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := make([]imageSequenceResponse, 0, len(sequences))
		for _, seq := range sequences {
			resp = append(resp, sequenceToResponse(seq))
		}
		writeJSON(w, http.StatusOK, map[string]any{"sequences": resp})
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
			name = "Untitled Sequence"
		}
		sequence := &repo.ImageSequence{
			ProjectID:   projectID,
			Name:        name,
			Description: strings.TrimSpace(req.Description),
		}
		if err := s.repos.ImageSequences.Create(sequence); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"sequence": sequenceToResponse(sequence)})
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Image sequence collection ---------------------------------------------------

func (s *Server) handleSequenceRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/image-sequences/")
	if path == "" || path == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(path, "/")
	sequenceID := parts[0]
	if sequenceID == "" {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 1 {
		s.handleSequenceItem(w, r, sequenceID)
		return
	}
	if parts[1] != "items" {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 2 {
		s.handleSequenceItems(w, r, sequenceID)
		return
	}
	position, err := strconv.Atoi(parts[2])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid item position")
		return
	}
	s.handleSequenceItemDeletion(w, r, sequenceID, position)
}

func (s *Server) handleSequenceItem(w http.ResponseWriter, r *http.Request, sequenceID string) {
	switch r.Method {
	case http.MethodGet:
		sequence, err := s.repos.ImageSequences.Get(sequenceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items, err := s.repos.ImageSequences.ListItems(sequenceID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"sequence": sequenceToResponse(sequence),
			"items":    sequenceItemsToResponse(items),
		})
	case http.MethodPatch, http.MethodPut:
		var req struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		sequence, err := s.repos.ImageSequences.Get(sequenceID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				name = "Untitled Sequence"
			}
			sequence.Name = name
		}
		if req.Description != nil {
			sequence.Description = strings.TrimSpace(*req.Description)
		}
		sequence.UpdatedAt = time.Now().UTC()
		if err := s.repos.ImageSequences.Update(sequence); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"sequence": sequenceToResponse(sequence)})
	case http.MethodDelete:
		if err := s.repos.ImageSequences.Delete(sequenceID); err != nil {
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

// Sequence items --------------------------------------------------------------

func (s *Server) handleSequenceItems(w http.ResponseWriter, r *http.Request, sequenceID string) {
	sequence, err := s.repos.ImageSequences.Get(sequenceID)
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
		items, err := s.repos.ImageSequences.ListItems(sequenceID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": sequenceItemsToResponse(items)})
	case http.MethodPost:
		var req struct {
			AssetID *string `json:"asset_id"`
			IsGap   bool    `json:"is_gap"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		if !req.IsGap {
			if req.AssetID == nil || *req.AssetID == "" {
				respondError(w, http.StatusBadRequest, "asset_id required when is_gap is false")
				return
			}
			asset, err := s.repos.Assets.Get(*req.AssetID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					respondError(w, http.StatusBadRequest, "asset not found")
					return
				}
				respondError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if asset.ProjectID != sequence.ProjectID {
				respondError(w, http.StatusBadRequest, "asset does not belong to sequence project")
				return
			}
		}
		item := &repo.ImageSequenceItem{
			SequenceID: sequenceID,
			AssetID:    req.AssetID,
			IsGap:      req.IsGap,
		}
		if err := s.repos.ImageSequences.AddItem(item); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items, err := s.repos.ImageSequences.ListItems(sequenceID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"items": sequenceItemsToResponse(items)})
	case http.MethodPut:
		var req struct {
			Items []struct {
				AssetID *string `json:"asset_id"`
				IsGap   bool    `json:"is_gap"`
			} `json:"items"`
		}
		if err := decodeJSON(r, &req); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		replacements := make([]*repo.ImageSequenceItem, 0, len(req.Items))
		for idx, item := range req.Items {
			if !item.IsGap {
				if item.AssetID == nil || *item.AssetID == "" {
					respondError(w, http.StatusBadRequest, fmt.Sprintf("asset_id required for item %d", idx))
					return
				}
				asset, err := s.repos.Assets.Get(*item.AssetID)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						respondError(w, http.StatusBadRequest, fmt.Sprintf("asset %s not found", *item.AssetID))
						return
					}
					respondError(w, http.StatusInternalServerError, err.Error())
					return
				}
				if asset.ProjectID != sequence.ProjectID {
					respondError(w, http.StatusBadRequest, "asset does not belong to sequence project")
					return
				}
			}
			replacements = append(replacements, &repo.ImageSequenceItem{
				SequenceID: sequenceID,
				AssetID:    item.AssetID,
				IsGap:      item.IsGap,
			})
		}
		if err := s.repos.ImageSequences.ReplaceItems(sequenceID, replacements); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items, err := s.repos.ImageSequences.ListItems(sequenceID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": sequenceItemsToResponse(items)})
	default:
		w.Header().Set("Allow", "GET, POST, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSequenceItemDeletion(w http.ResponseWriter, r *http.Request, sequenceID string, position int) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := s.repos.ImageSequences.DeleteItem(sequenceID, position); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
