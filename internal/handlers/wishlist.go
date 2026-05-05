package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/middleware"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/services"
)

var wishlistService *services.WishlistService

func SetWishlistService(svc *services.WishlistService) {
	wishlistService = svc
}

// WishlistCollectionHandler handles POST/GET on /api/wishlist.
// @Summary      Create or list wishlist items
// @Description  POST: Add game to wishlist | GET: List authenticated user's wishlist
// @Tags         wishlist
// @Accept       json
// @Produce      json
// @Param        request  body      models.WishlistCreateRequest  false  "Create request (POST only)"
// @Param        limit    query     int                           false  "Page size"  default(20)
// @Param        offset   query     int                           false  "Page offset"  default(0)
// @Success      201      {object}  models.WishlistItem
// @Success      200      {object}  models.WishlistListResponse
// @Failure      400      {object}  models.APIError
// @Failure      401      {object}  models.APIError
// @Failure      409      {object}  models.APIError
// @Failure      500      {object}  models.APIError
// @Router       /api/wishlist [post]
// @Router       /api/wishlist [get]
func WishlistCollectionHandler(w http.ResponseWriter, r *http.Request) {
	if wishlistService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Wishlist service not initialized"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, models.APIError{Error: "Unauthorized"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		var req models.WishlistCreateRequest
		if err := decodeJSONBody(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request body"})
			return
		}
		item, err := wishlistService.AddItem(r.Context(), userID, req)
		if err != nil {
			writeServiceError(w, err, "Failed to add wishlist item")
			return
		}
		writeJSON(w, http.StatusCreated, item)
	case http.MethodGet:
		limit := parseQueryInt(r.URL.Query().Get("limit"), 20)
		offset := parseQueryInt(r.URL.Query().Get("offset"), 0)
		response, err := wishlistService.ListItems(r.Context(), userID, limit, offset)
		if err != nil {
			writeServiceError(w, err, "Failed to list wishlist")
			return
		}
		writeJSON(w, http.StatusOK, response)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
	}
}

// WishlistItemHandler handles PATCH /api/wishlist/{id}/threshold and DELETE /api/wishlist/{id}.
// @Summary      Update or remove wishlist item
// @Description  PATCH: Update target price for wishlist item | DELETE: Remove item from wishlist
// @Tags         wishlist
// @Accept       json
// @Produce      json
// @Param        id       path      int64                         true   "Wishlist item ID"
// @Param        request  body      models.WishlistUpdateRequest  false  "Update request (PATCH only)"
// @Success      200      {object}  models.WishlistItem
// @Success      204
// @Failure      400      {object}  models.APIError
// @Failure      401      {object}  models.APIError
// @Failure      404      {object}  models.APIError
// @Failure      500      {object}  models.APIError
// @Router       /api/wishlist/{id}/threshold [patch]
// @Router       /api/wishlist/{id} [delete]
func WishlistItemHandler(w http.ResponseWriter, r *http.Request) {
	if wishlistService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Wishlist service not initialized"})
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, models.APIError{Error: "Unauthorized"})
		return
	}

	const prefix = "/api/wishlist/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid wishlist path"})
		return
	}

	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, prefix), "/")
	if r.Method == http.MethodPatch {
		if !strings.HasSuffix(rest, "/threshold") {
			writeJSON(w, http.StatusNotFound, models.APIError{Error: "Threshold route not found"})
			return
		}
		rest = strings.TrimSuffix(rest, "/threshold")
	}
	wishlistID, err := strconv.ParseInt(rest, 10, 64)
	if err != nil || wishlistID <= 0 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid wishlist id"})
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var req models.WishlistUpdateRequest
		if err := decodeJSONBody(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid request body"})
			return
		}
		item, err := wishlistService.UpdateItem(r.Context(), userID, wishlistID, req)
		if err != nil {
			writeServiceError(w, err, "Failed to update wishlist item")
			return
		}
		writeJSON(w, http.StatusOK, item)
	case http.MethodDelete:
		if err := wishlistService.DeleteItem(r.Context(), userID, wishlistID); err != nil {
			writeServiceError(w, err, "Failed to delete wishlist item")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
	}
}

// BuyAdviceHandler returns buy-now vs wait recommendation for a game.
// @Summary      Get buy timing advice
// @Description  Returns intelligent buy-now or wait recommendation based on 90-day price history
// @Tags         prices
// @Produce      json
// @Param        id   path  int64  true  "Game ID"
// @Success      200  {object}  models.BuyAdviceResponse
// @Failure      400  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/games/{id}/buy-advice [get]
func BuyAdviceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if gamesService == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Games service not initialized"})
		return
	}

	const prefix = "/api/games/"
	const suffix = "/buy-advice"
	if !strings.HasPrefix(r.URL.Path, prefix) || !strings.HasSuffix(r.URL.Path, suffix) {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid advice path"})
		return
	}

	middle := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), suffix)
	gameID, err := strconv.ParseInt(strings.Trim(middle, "/"), 10, 64)
	if err != nil || gameID <= 0 {
		writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid game id"})
		return
	}

	advice, err := gamesService.GetBuyAdvice(r.Context(), gameID)
	if err != nil {
		writeServiceError(w, err, "Failed to calculate buy advice")
		return
	}

	writeJSON(w, http.StatusOK, advice)
}
