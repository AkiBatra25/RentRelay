package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

type AgreementHandler struct {
	client rentrelaypb.AgreementServiceClient
}

func NewAgreementHandler(client rentrelaypb.AgreementServiceClient) *AgreementHandler {
	return &AgreementHandler{client: client}
}

func (h *AgreementHandler) CreateAgreement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var body struct {
		TenantID      string  `json:"tenant_id"`
		LandlordID    string  `json:"landlord_id"`
		PropertyID    string  `json:"property_id"`
		MonthlyRent   float64 `json:"monthly_rent"`
		DepositAmount float64 `json:"deposit_amount"`
		LeaseMonths   int32   `json:"lease_months"`
		NoticeDays    int32   `json:"notice_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	resp, err := h.client.CreateAgreement(r.Context(), &rentrelaypb.CreateAgreementRequest{
		TenantId: body.TenantID, LandlordId: body.LandlordID,
		PropertyId: body.PropertyID, MonthlyRent: body.MonthlyRent,
		DepositAmount: body.DepositAmount, LeaseMonths: body.LeaseMonths,
		NoticeDays: body.NoticeDays,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, grpcErrMsg(err))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"agreement_id": resp.AgreementId,
		"state":        resp.State.String(),
		"monthly_rent": resp.MonthlyRent,
		"deposit":      resp.DepositAmount,
	})
}

func (h *AgreementHandler) GetAgreement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	agreementID := strings.TrimPrefix(r.URL.Path, "/api/agreements/")
	if agreementID == "" {
		writeError(w, http.StatusBadRequest, "agreement_id is required in URL")
		return
	}
	resp, err := h.client.GetAgreement(r.Context(), &rentrelaypb.AgreementActionRequest{
		AgreementId: agreementID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, grpcErrMsg(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"agreement_id": resp.AgreementId,
		"state":        resp.State.String(),
		"tenant_id":    resp.TenantId,
		"landlord_id":  resp.LandlordId,
		"property_id":  resp.PropertyId,
		"monthly_rent": resp.MonthlyRent,
		"deposit":      resp.DepositAmount,
		"deposit_held": resp.DepositHeld,
	})
}

func (h *AgreementHandler) SignAgreement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/agreements/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "sign" {
		writeError(w, http.StatusBadRequest, "URL must be /api/agreements/{id}/sign")
		return
	}
	agreementID := parts[0]
	var body struct {
		SignerID      string `json:"signer_id"`
		SignatureHash string `json:"signature_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	resp, err := h.client.SignAgreement(r.Context(), &rentrelaypb.SignAgreementRequest{
		AgreementId: agreementID, SignerId: body.SignerID,
		SignatureHash: body.SignatureHash,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, grpcErrMsg(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"agreement_id": resp.AgreementId,
		"state":        resp.State.String(),
		"signatures":   len(resp.Signatures),
	})
}
