package usecase

import (
	"context"
	"strings"

	"github.com/newfeed/community-news/services/user-service/internal/profile/domain"
)

type Repository interface {
	UpsertProfile(ctx context.Context, profile domain.Profile) error
	GetProfile(ctx context.Context, userID string) (domain.Profile, error)
	UpdateAvatar(ctx context.Context, userID, objectID string) error
	UpdateCover(ctx context.Context, userID, objectID string) error
	FollowUser(ctx context.Context, followerID, followedID string) error
	FollowTopic(ctx context.Context, userID, topic string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (Service) FollowTopic(profile domain.Profile, topic string) domain.Profile {
	profile.Topics = append(profile.Topics, topic)
	return profile
}

func (s Service) UpsertProfile(ctx context.Context, profile domain.Profile) error {
	profile.DisplayName = strings.TrimSpace(profile.DisplayName)
	profile.Bio = strings.TrimSpace(profile.Bio)
	profile.Headline = strings.TrimSpace(profile.Headline)
	profile.Education = strings.TrimSpace(profile.Education)
	profile.Occupation = strings.TrimSpace(profile.Occupation)
	profile.Location = strings.TrimSpace(profile.Location)
	profile.WebsiteURL = strings.TrimSpace(profile.WebsiteURL)
	return s.repo.UpsertProfile(ctx, profile)
}

func (s Service) GetProfile(ctx context.Context, userID string) (domain.Profile, error) {
	return s.repo.GetProfile(ctx, userID)
}

func (s Service) UpdateAvatar(ctx context.Context, userID, objectID string) error {
	return s.repo.UpdateAvatar(ctx, userID, objectID)
}

func (s Service) UpdateCover(ctx context.Context, userID, objectID string) error {
	return s.repo.UpdateCover(ctx, userID, objectID)
}

func (s Service) FollowUser(ctx context.Context, followerID, followedID string) error {
	return s.repo.FollowUser(ctx, followerID, followedID)
}

func (s Service) FollowTopicByUser(ctx context.Context, userID, topic string) error {
	return s.repo.FollowTopic(ctx, userID, strings.ToLower(strings.TrimSpace(topic)))
}
