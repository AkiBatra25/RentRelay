package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
)

type UserHandler struct {
	client rentrelaypb.UserServiceClient
}

func NewUserHandler(client rentrelaypb.UserServiceClient) *UserHandler {
	return &UserHandler{client: client}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	role := rentrelaypb.UserRole_ROLE_TENANT
	if strings.ToLower(body.Role) == "landlord" {
		role = rentrelaypb.UserRole_ROLE_LANDLORD
	}
	resp, err := h.client.Register(r.Context(), &rentrelaypb.RegisterRequest{
		Name: body.Name, Email: body.Email, Phone: body.Phone,
		Password: body.Password, Role: role,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, grpcErrMsg(err))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"user_id": resp.User.UserId,
		"email":   resp.User.Email,
		"name":    resp.User.Name,
		"role":    resp.User.Role.String(),
		"token":   resp.Token,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	resp, err := h.client.Login(r.Context(), &rentrelaypb.LoginRequest{
		Email: body.Email, Password: body.Password,
	})
	if err != nil {
		writeError(w, http.StatusUnauthorized, grpcErrMsg(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token":   resp.Token,
		"user_id": resp.User.UserId,
		"email":   resp.User.Email,
	})
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	userID := strings.TrimPrefix(r.URL.Path, "/api/users/")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required in URL")
		return
	}
	user, err := h.client.GetUser(r.Context(), &rentrelaypb.GetUserRequest{UserId: userID})
	if err != nil {
		writeError(w, http.StatusNotFound, grpcErrMsg(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id": user.UserId,
		"email":   user.Email,
		"name":    user.Name,
		"role":    user.Role.String(),
		"kyc":     user.KycVerified,
	})
}
