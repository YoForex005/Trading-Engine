package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// PrintPreferences represents user's chart print settings
type PrintPreferences struct {
	PageSize           string  `json:"pageSize"`
	Orientation        string  `json:"orientation"`
	ColorMode          string  `json:"colorMode"`
	IncludeHeader      bool    `json:"includeHeader"`
	IncludeFooter      bool    `json:"includeFooter"`
	IncludeGrid        bool    `json:"includeGrid"`
	IncludeIndicators  bool    `json:"includeIndicators"`
	IncludeDrawings    bool    `json:"includeDrawings"`
	MarginTop          float64 `json:"marginTop"`
	MarginBottom       float64 `json:"marginBottom"`
	MarginLeft         float64 `json:"marginLeft"`
	MarginRight        float64 `json:"marginRight"`
	ScaleToFit         bool    `json:"scaleToFit"`
	HeaderText         string  `json:"headerText,omitempty"`
	FooterText         string  `json:"footerText,omitempty"`
}

// PrintPreferencesStore manages print preferences per account
type PrintPreferencesStore struct {
	mu          sync.RWMutex
	preferences map[string]*PrintPreferences // accountID -> preferences
}

// NewPrintPreferencesStore creates a new print preferences store
func NewPrintPreferencesStore() *PrintPreferencesStore {
	return &PrintPreferencesStore{
		preferences: make(map[string]*PrintPreferences),
	}
}

// GetDefault returns default print preferences
func (s *PrintPreferencesStore) GetDefault() *PrintPreferences {
	return &PrintPreferences{
		PageSize:          "A4",
		Orientation:       "landscape",
		ColorMode:         "color",
		IncludeHeader:     true,
		IncludeFooter:     true,
		IncludeGrid:       true,
		IncludeIndicators: true,
		IncludeDrawings:   true,
		MarginTop:         20,
		MarginBottom:      20,
		MarginLeft:        20,
		MarginRight:       20,
		ScaleToFit:        true,
	}
}

// Get retrieves print preferences for an account
func (s *PrintPreferencesStore) Get(accountID string) *PrintPreferences {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if prefs, ok := s.preferences[accountID]; ok {
		return prefs
	}

	return s.GetDefault()
}

// Set saves print preferences for an account
func (s *PrintPreferencesStore) Set(accountID string, prefs *PrintPreferences) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.preferences[accountID] = prefs
}

// Delete removes print preferences for an account
func (s *PrintPreferencesStore) Delete(accountID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.preferences, accountID)
}

// HandleGetPrintPreferences handles GET /api/accounts/{accountId}/print-preferences
func (s *PrintPreferencesStore) HandleGetPrintPreferences(w http.ResponseWriter, r *http.Request) {
	// Extract accountId from URL path: /api/accounts/{accountId}/print-preferences
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	accountID := pathParts[2] // accounts/{accountId}/print-preferences

	if accountID == "" {
		http.Error(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	prefs := s.Get(accountID)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(prefs); err != nil {
		http.Error(w, "Failed to encode preferences", http.StatusInternalServerError)
		return
	}
}

// HandleSavePrintPreferences handles POST /api/accounts/{accountId}/print-preferences
func (s *PrintPreferencesStore) HandleSavePrintPreferences(w http.ResponseWriter, r *http.Request) {
	// Extract accountId from URL path: /api/accounts/{accountId}/print-preferences
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	accountID := pathParts[2] // accounts/{accountId}/print-preferences

	if accountID == "" {
		http.Error(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	var prefs PrintPreferences
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate preferences
	if err := validatePrintPreferences(&prefs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.Set(accountID, &prefs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Print preferences saved successfully",
	})
}

// HandleDeletePrintPreferences handles DELETE /api/accounts/{accountId}/print-preferences
func (s *PrintPreferencesStore) HandleDeletePrintPreferences(w http.ResponseWriter, r *http.Request) {
	// Extract accountId from URL path: /api/accounts/{accountId}/print-preferences
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	accountID := pathParts[2] // accounts/{accountId}/print-preferences

	if accountID == "" {
		http.Error(w, "Account ID is required", http.StatusBadRequest)
		return
	}

	s.Delete(accountID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Print preferences deleted successfully",
	})
}

// validatePrintPreferences validates print preferences
func validatePrintPreferences(prefs *PrintPreferences) error {
	// Validate page size
	validPageSizes := map[string]bool{
		"A4": true, "Letter": true, "Legal": true, "A3": true,
	}
	if !validPageSizes[prefs.PageSize] {
		return http.ErrAbortHandler // Invalid page size
	}

	// Validate orientation
	if prefs.Orientation != "portrait" && prefs.Orientation != "landscape" {
		return http.ErrAbortHandler // Invalid orientation
	}

	// Validate color mode
	if prefs.ColorMode != "color" && prefs.ColorMode != "grayscale" {
		return http.ErrAbortHandler // Invalid color mode
	}

	// Validate margins (0-50)
	if prefs.MarginTop < 0 || prefs.MarginTop > 50 ||
		prefs.MarginBottom < 0 || prefs.MarginBottom > 50 ||
		prefs.MarginLeft < 0 || prefs.MarginLeft > 50 ||
		prefs.MarginRight < 0 || prefs.MarginRight > 50 {
		return http.ErrAbortHandler // Invalid margins
	}

	return nil
}
