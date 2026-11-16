package dashboard

import (
	"encoding/json"
	"net/http"

	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/dashboard"
)

type DashboardHandler struct {
	service dashboard.DashboardService
}

func NewDashboardHandler(s dashboard.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: s}
}

func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var resp any

	data, err := h.service.GetDashboard(r.Context())
	if err != nil {
		resp = ErrorResponse{
			StatusCode: 500,
			Message:    "Internal server error " + err.Error(),
		}
	} else {
		resp = GetResponsePage[dashboard.Dashboard]{
			StatusCode: 200,
			Message:    "Dashboard fetched successfully",
			Data:       *data,
		}
	}
	json.NewEncoder(w).Encode(resp)
}
