package handlers

import (
	"net/http"
	"strconv"

	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/models"
	"github.com/tanmaybhardwaj2004/dropsandgrinds/internal/repositories"
)

var indianPaymentRepo *repositories.IndianPaymentRepository

// SetIndianPaymentRepository wires the Indian payment repository into HTTP handlers at startup.
func SetIndianPaymentRepository(repo *repositories.IndianPaymentRepository) {
	indianPaymentRepo = repo
}

// IndianOffersHandler handles GET on /api/indian-offers.
// @Summary      Get Indian payment offers
// @Description  Get all active Indian payment offers for game stores
// @Tags         indian-offers
// @Produce      json
// @Param        store_id query int64 false "Filter by store ID"
// @Param        provider query string false "Filter by payment provider (e.g., phonepe, gpay, paytm)"
// @Success      200 {object} models.IndianOffersResponse
// @Failure      500 {object} models.APIError
// @Router       /api/indian-offers [get]
func IndianOffersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "Method not allowed"})
		return
	}
	if indianPaymentRepo == nil {
		writeJSON(w, http.StatusInternalServerError, models.APIError{Error: "Indian payment repository not initialized"})
		return
	}

	var offers []models.IndianPaymentOffer
	var err error

	storeIDStr := r.URL.Query().Get("store_id")
	provider := r.URL.Query().Get("provider")

	if storeIDStr != "" {
		storeID, parseErr := strconv.ParseInt(storeIDStr, 10, 64)
		if parseErr != nil {
			writeJSON(w, http.StatusBadRequest, models.APIError{Error: "Invalid store_id"})
			return
		}
		offers, err = indianPaymentRepo.GetOffersByStore(r.Context(), storeID)
	} else if provider != "" {
		offers, err = indianPaymentRepo.GetOffersByProvider(r.Context(), provider)
	} else {
		offers, err = indianPaymentRepo.GetActiveOffers(r.Context())
	}

	if err != nil {
		writeServiceError(w, err, "Failed to get Indian payment offers")
		return
	}

	response := models.IndianOffersResponse{
		Offers: offers,
		Total:  len(offers),
	}

	writeJSON(w, http.StatusOK, response)
}
