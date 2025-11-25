package service

import (
	"context"
	"log/slog"

	"avito/internal/domain"
	"avito/internal/repository"
)

type StatisticsService struct {
	statsRepo repository.StatisticsRepository
}

func NewStatisticsService(statsRepo repository.StatisticsRepository) *StatisticsService {
	return &StatisticsService{
		statsRepo: statsRepo,
	}
}

func (s *StatisticsService) GetStatistics(ctx context.Context) (*domain.Statistics, error) {
	stats := &domain.Statistics{}

	slog.Info("saving user stats")
	assignmentsByUser, err := s.statsRepo.GetAssignmentsByUser(ctx)
	if err != nil {
		return nil, err
	}
	stats.AssignmentsByUser = assignmentsByUser

	for _, stat := range assignmentsByUser {
		stats.TotalAssignments += stat.Count
	}

	slog.Info("found stats for assignments", "total", stats.TotalAssignments)

	assignmentsByPR, err := s.statsRepo.GetAssignmentsByPR(ctx)
	if err != nil {
		return nil, err
	}
	stats.AssignmentsByPR = assignmentsByPR

	totalPRs, err := s.statsRepo.GetTotalPRs(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalPRs = totalPRs

	activeUsers, err := s.statsRepo.GetActiveUsersCount(ctx)
	if err != nil {
		return nil, err
	}
	stats.ActiveUsers = activeUsers

	teams, err := s.statsRepo.GetTeamsCount(ctx)
	if err != nil {
		return nil, err
	}
	stats.Teams = teams

	return stats, nil
}
