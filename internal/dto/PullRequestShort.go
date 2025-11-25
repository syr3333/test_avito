package dto

import "avito/internal/domain"

type PullRequestShort struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}

func PullRequestShortFromDomain(pr domain.PullRequestShort) PullRequestShort {
	return PullRequestShort{
		ID:       pr.ID,
		Name:     pr.Name,
		AuthorID: pr.AuthorID,
		Status:   pr.Status,
	}
}

func PullRequestsShortFromDomain(prs []domain.PullRequestShort) []PullRequestShort {
	result := make([]PullRequestShort, len(prs))
	for i, pr := range prs {
		result[i] = PullRequestShortFromDomain(pr)
	}
	return result
}
