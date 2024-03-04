package services

import (
	"fmt"
	"myproject/api-gateway/config"
	pbu "myproject/api-gateway/genproto/user-service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

type IServiceManager interface {
	UserService() pbu.UserServiceClient
}

type serviceManager struct {
	userService pbu.UserServiceClient
}

func (s *serviceManager) UserService() pbu.UserServiceClient {
	return s.userService
}

func NewServiceManager(cfg config.Config) (IServiceManager, error) {
	resolver.SetDefaultScheme("dns")

	connUser, err := grpc.Dial(
		fmt.Sprintf("%s:%d", cfg.UserServiceHost, cfg.UserServicePort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("user service dial error, %s:%s:%v", cfg.UserServiceHost, cfg.UserServicePort, err)
	}

	return &serviceManager{userService: pbu.NewUserServiceClient(connUser)}, nil
}
