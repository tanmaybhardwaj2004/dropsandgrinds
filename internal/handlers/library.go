package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

var libraryService *services.LibraryService

// SetLibraryService sets the library service
func SetLibraryService(svc *services.LibraryService) {
	libraryService = svc
}

// LibraryImportHandler handles POST /api/library/import
// @Summary      Import Steam library
// @Description  Imports owned games from Steam using SteamID. Requires authentication.
// @Tags         library
// @Accept       json
// @Produce      json
// @Param        request  body      models.LibraryImportRequest  true  "Import request"
// @Success      200  {object}  services.ImportResult
// @Failure      400  {object}  models.APIError
// @Failure      401  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/library/import [post]
func LibraryImportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if libraryService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Library service not initialized"})
		return
	}

	userID := r.Context().Value("user_id")
	if userID == nil {
		writeJSON(w, http.StatusUnauthorized, models.APIError{Error: "Unauthorized"})
		return
	}

	var req models.LibraryImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request body"})
		return
	}

	// Validate SteamID
	if req.SteamID == "" {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "SteamID is required"})
		return
	}

	result, err := libraryService.ImportLibrary(r.Context(), userID.(int64), req.SteamID)
	if err != nil {
		writeServiceError(w, err, "Failed to import library")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// LibraryListHandler handles GET /api/library
// @Summary      Get owned games
// @Description  Returns list of owned game IDs for the authenticated user
// @Tags         library
// @Produce      json
// @Success      200  {object}  models.LibraryListResponse
// @Failure      401  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/library [get]
func LibraryListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if libraryService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Library service not initialized"})
		return
	}

	userID := r.Context().Value("user_id")
	if userID == nil {
		writeJSON(w, http.StatusUnauthorized, models.APIError{Error: "Unauthorized"})
		return
	}

	gameIDs, err := libraryService.GetLibrary(r.Context(), userID.(int64))
	if err != nil {
		writeServiceError(w, err, "Failed to get library")
		return
	}

	response := models.LibraryListResponse{
		GameIDs: gameIDs,
		Count:   len(gameIDs),
	}

	writeJSON(w, http.StatusOK, response)
}
