package handlers

import (
	"net/http"

	"avito/internal/domain"
	"avito/internal/dto"
	"avito/internal/service"
)

type StatisticsHandler struct {
	statsService *service.StatisticsService
}

func NewStatisticsHandler(statsService *service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{
		statsService: statsService,
	}
}

func (h *StatisticsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.statsService.GetStatistics(ctx)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, domain.ErrCodeInternalError, err.Error())
		return
	}

	response := dto.StatisticsFromDomain(stats)
	WriteJSON(w, http.StatusOK, response)
}
