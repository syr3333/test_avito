package dto

import "avito/internal/domain"

type AssignmentStat struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
}

type StatisticsResponse struct {
	AssignmentsByUser []AssignmentStat `json:"assignments_by_user"`
	AssignmentsByPR   []AssignmentStat `json:"assignments_by_pr"`
	TotalPRs          int              `json:"total_prs"`
	TotalAssignments  int              `json:"total_assignments"`
	ActiveUsers       int              `json:"active_users"`
	Teams             int              `json:"teams"`
}

func StatisticsFromDomain(stats *domain.Statistics) *StatisticsResponse {
	response := &StatisticsResponse{
		TotalPRs:         stats.TotalPRs,
		TotalAssignments: stats.TotalAssignments,
		ActiveUsers:      stats.ActiveUsers,
		Teams:            stats.Teams,
	}

	response.AssignmentsByUser = make([]AssignmentStat, len(stats.AssignmentsByUser))
	for i, stat := range stats.AssignmentsByUser {
		response.AssignmentsByUser[i] = AssignmentStat{
			ID:    stat.ID,
			Count: stat.Count,
		}
	}

	response.AssignmentsByPR = make([]AssignmentStat, len(stats.AssignmentsByPR))
	for i, stat := range stats.AssignmentsByPR {
		response.AssignmentsByPR[i] = AssignmentStat{
			ID:    stat.ID,
			Count: stat.Count,
		}
	}

	return response
}
