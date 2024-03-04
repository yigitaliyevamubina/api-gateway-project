package v1

import (
	"errors"
	"fmt"
	"github.com/casbin/casbin/v2"
	"myproject/api-gateway/api/handlers/tokens"
	"myproject/api-gateway/api/models"
	"myproject/api-gateway/config"
	"myproject/api-gateway/pkg/logger"
	grpcClient "myproject/api-gateway/services"
	"myproject/api-gateway/storage/repo"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type handlerV1 struct {
	inMemoryStorage repo.InMemoryStorageI
	log             logger.Logger
	serviceManager  grpcClient.IServiceManager
	cfg             config.Config
	jwtHandler      tokens.JWTHandler
	casbin          *casbin.Enforcer
}

type HandlerV1Config struct {
	InMemoryStorage repo.InMemoryStorageI
	Log             logger.Logger
	ServiceManager  grpcClient.IServiceManager
	Cfg             config.Config
	JwtHandler      tokens.JWTHandler
	Casbin          *casbin.Enforcer
}

func New(h *HandlerV1Config) *handlerV1 {
	return &handlerV1{
		inMemoryStorage: h.InMemoryStorage,
		log:             h.Log,
		serviceManager:  h.ServiceManager,
		cfg:             h.Cfg,
		jwtHandler:      h.JwtHandler,
		casbin:          h.Casbin,
	}
}

const (
	ErrorCodeInvalidURL          = "INVALID_URL"
	ErrorCodeInvalidJSON         = "INVALID_JSON"
	ErrorCodeInvalidParams       = "INVALID_PARAMS"
	ErrorCodeInternalServerError = "INTERNAL_SERVER_ERROR"
	ErrorCodeUnauthorized        = "UNAUTHORIZED"
	ErrorCodeAlreadyExists       = "ALREADY_EXISTS"
	ErrorCodeNotFound            = "NOT_FOUND"
	ErrorCodeInvalidCode         = "INVALID_CODE"
	ErrorBadRequest              = "BAD_REQUEST"
	ErrorInvalidCredentials      = "INVALID_CREDENTIALS"
	StatusMethodNotAllowed       = "METHOD_NOT_ALLOWED"
	ErrorValidationError         = "VALIDATION_ERROR"
)

func ParsePageQueryParam(c *gin.Context) (int, error) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		return 0, err
	}
	if page < 0 {
		return 0, fmt.Errorf("page should be a positive number")
	}
	if page == 0 {
		return 1, nil
	}

	return page, nil
}

func ParseLimitQueryParam(c *gin.Context) (int, error) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		return 0, err
	}
	if limit < 0 {
		return 0, errors.New("page size should be a positive number")
	}

	if limit == 0 {
		return 10, nil
	}

	return limit, nil
}

func handleBadRequestErrWithMessage(c *gin.Context, log logger.Logger, err error, status string) bool {
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ResponseError{
			Error: models.StandardErrorModel{
				Status:  status,
				Message: err.Error(),
			},
		})
		log.Error(err.Error(), logger.Error(err))
		return true
	}
	return false
}

func handleInternalServerErrorWithMessage(c *gin.Context, log logger.Logger, err error, message string) bool {
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ResponseError{
			Error: models.ServerError{
				Status:  ErrorCodeInternalServerError,
				Message: "Sorry, try again",
			},
		})
		log.Error(message, logger.Error(err))
		return true
	}

	return false
}
