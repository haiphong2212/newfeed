package grpc

import (
	"context"

	userv1 "github.com/newfeed/community-news/gen/user/v1"
	"github.com/newfeed/community-news/services/user-service/internal/profile/domain"
	"github.com/newfeed/community-news/services/user-service/internal/profile/usecase"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	profiles usecase.Service
}

func NewServer(profiles usecase.Service) *Server {
	return &Server{profiles: profiles}
}

func (s *Server) UpsertProfile(ctx context.Context, req *userv1.UpsertProfileRequest) (*userv1.UpsertProfileResponse, error) {
	err := s.profiles.UpsertProfile(ctx, domain.Profile{
		UserID:         req.GetUserId(),
		DisplayName:    req.GetDisplayName(),
		Bio:            req.GetBio(),
		Headline:       req.GetHeadline(),
		Education:      req.GetEducation(),
		Occupation:     req.GetOccupation(),
		Location:       req.GetLocation(),
		WebsiteURL:     req.GetWebsiteUrl(),
		AvatarObjectID: req.GetAvatarObjectId(),
		CoverObjectID:  req.GetCoverObjectId(),
	})
	return &userv1.UpsertProfileResponse{}, err
}

func (s *Server) FollowUser(ctx context.Context, req *userv1.FollowUserRequest) (*userv1.FollowUserResponse, error) {
	return &userv1.FollowUserResponse{}, s.profiles.FollowUser(ctx, req.GetFollowerId(), req.GetFollowedId())
}

func (s *Server) FollowTopic(ctx context.Context, req *userv1.FollowTopicRequest) (*userv1.FollowTopicResponse, error) {
	return &userv1.FollowTopicResponse{}, s.profiles.FollowTopicByUser(ctx, req.GetUserId(), req.GetTopic())
}
