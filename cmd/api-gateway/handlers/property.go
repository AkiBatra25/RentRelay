package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

type PropertyHandler struct {
	client rentrelaypb.PropertyServiceClient
}

func NewPropertyHandler(client rentrelaypb.PropertyServiceClient) *PropertyHandler {
	return &PropertyHandler{client: client}
}

// RegisterProperty handles POST /api/properties
// Body: { "landlord_id": "...", "title": "2BHK HSR", "city": "Bengaluru", ... }
func (h *PropertyHandler) RegisterProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}

	var body struct {
		LandlordID  string   `json:"landlord_id"`
		Title       string   `json:"title"`
		Address     string   `json:"address"`
		City        string   `json:"city"`
		Zone        string   `json:"zone"`
		Bedrooms    int32    `json:"bedrooms"`
		RentMonthly float64  `json:"rent_monthly"`
		DepositAmt  float64  `json:"deposit_amt"`
		Furnishing  string   `json:"furnishing"` // "furnished", "semi_furnished", "unfurnished"
		Amenities   []string `json:"amenities"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	furnishing := rentrelaypb.FurnishingType_UNFURNISHED
	switch strings.ToLower(body.Furnishing) {
	case "furnished":
		furnishing = rentrelaypb.FurnishingType_FULLY_FURNISHED
	case "semi_furnished":
		furnishing = rentrelaypb.FurnishingType_SEMI_FURNISHED
	}

	resp, err := h.client.RegisterProperty(r.Context(), &rentrelaypb.RegisterPropertyRequest{
		LandlordId:  body.LandlordID,
		Title:       body.Title,
		Address:     body.Address,
		City:        body.City,
		Zone:        body.Zone,
		Bedrooms:    body.Bedrooms,
		RentMonthly: body.RentMonthly,
		DepositAmt:  body.DepositAmt,
		Furnishing:  furnishing,
		Amenities:   body.Amenities,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, grpcErrMsg(err))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"property_id": resp.PropertyId,
		"title":       resp.Title,
		"city":        resp.City,
		"rent":        resp.RentMonthly,
	})
}

// SearchProperties handles GET /api/properties/search?city=Bengaluru&max_rent=30000&bedrooms=2
func (h *PropertyHandler) SearchProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}

	// Read query parameters from the URL
	q := r.URL.Query()
	city := q.Get("city")
	zone := q.Get("zone")
	maxRent, _ := strconv.ParseFloat(q.Get("max_rent"), 64)
	bedrooms, _ := strconv.ParseInt(q.Get("bedrooms"), 10, 32)

	resp, err := h.client.SearchProperties(r.Context(), &rentrelaypb.SearchPropertiesRequest{
		City:        city,
		Zone:        zone,
		MaxRent:     maxRent,
		MinBedrooms: int32(bedrooms),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, grpcErrMsg(err))
		return
	}

	// Build a clean list of properties to return
	properties := make([]map[string]any, 0, len(resp.Properties))
	for _, p := range resp.Properties {
		properties = append(properties, map[string]any{
			"property_id": p.PropertyId,
			"title":       p.Title,
			"city":        p.City,
			"zone":        p.Zone,
			"bedrooms":    p.Bedrooms,
			"rent":        p.RentMonthly,
			"furnishing":  p.Furnishing.String(),
			"available":   p.IsAvailable,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"count":      len(properties),
		"properties": properties,
	})
}

// GetProperty handles GET /api/properties/{id}
func (h *PropertyHandler) GetProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}

	propertyID := strings.TrimPrefix(r.URL.Path, "/api/properties/")
	if propertyID == "" || propertyID == "search" {
		writeError(w, http.StatusBadRequest, "property_id is required in URL")
		return
	}

	resp, err := h.client.GetProperty(r.Context(), &rentrelaypb.GetPropertyRequest{PropertyId: propertyID})
	if err != nil {
		writeError(w, http.StatusNotFound, grpcErrMsg(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"property_id": resp.PropertyId,
		"title":       resp.Title,
		"city":        resp.City,
		"zone":        resp.Zone,
		"bedrooms":    resp.Bedrooms,
		"rent":        resp.RentMonthly,
		"deposit":     resp.DepositAmt,
		"furnishing":  resp.Furnishing.String(),
		"available":   resp.IsAvailable,
		"amenities":   resp.Amenities,
	})
}
