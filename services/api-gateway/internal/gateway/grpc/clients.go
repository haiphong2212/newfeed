package grpc

import (
	authv1 "github.com/newfeed/community-news/gen/auth/v1"
	newsv1 "github.com/newfeed/community-news/gen/news/v1"
	searchv1 "github.com/newfeed/community-news/gen/search/v1"
	userv1 "github.com/newfeed/community-news/gen/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Auth   authv1.AuthServiceClient
	News   newsv1.NewsServiceClient
	User   userv1.UserServiceClient
	Search searchv1.SearchServiceClient
	conns  []*grpc.ClientConn
}

func NewClients(targets map[string]string) (*Clients, error) {
	authConn, err := dial(targets["auth-service"])
	if err != nil {
		return nil, err
	}
	userConn, err := dial(targets["user-service"])
	if err != nil {
		return nil, err
	}
	newsConn, err := dial(targets["news-service"])
	if err != nil {
		return nil, err
	}
	searchConn, err := dial(targets["search-service"])
	if err != nil {
		return nil, err
	}
	return &Clients{
		Auth:   authv1.NewAuthServiceClient(authConn),
		User:   userv1.NewUserServiceClient(userConn),
		News:   newsv1.NewNewsServiceClient(newsConn),
		Search: searchv1.NewSearchServiceClient(searchConn),
		conns:  []*grpc.ClientConn{authConn, userConn, newsConn, searchConn},
	}, nil
}

func (c *Clients) Close() {
	for _, conn := range c.conns {
		_ = conn.Close()
	}
}

func dial(target string) (*grpc.ClientConn, error) {
	return grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
