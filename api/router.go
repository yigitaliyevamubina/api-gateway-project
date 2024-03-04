package api

import (
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/util"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "myproject/api-gateway/api/docs"
	"myproject/api-gateway/api/handlers/tokens"
	v1 "myproject/api-gateway/api/handlers/v1"
	"myproject/api-gateway/api/middleware"
	"myproject/api-gateway/config"
	"myproject/api-gateway/pkg/logger"
	"myproject/api-gateway/services"
	"myproject/api-gateway/storage/repo"
)

type Option struct {
	InMemory       repo.InMemoryStorageI
	Cfg            config.Config
	Logger         logger.Logger
	ServiceManager services.IServiceManager
}

// Constructor
// @Title Super Clinic
// @version 1.0
// @description api-gateway
// @host localhost:9090
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func New(option Option) *gin.Engine {
	psqlString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		option.Cfg.PostgresHost,
		option.Cfg.PostgresPort,
		option.Cfg.PostgresUser,
		option.Cfg.PostgresPassword,
		option.Cfg.PostgresDatabase)

	adapter, err := gormadapter.NewAdapter("postgres", psqlString, true)
	if err != nil {
		option.Logger.Fatal("error while creating a new casbin adapter\n", logger.Error(err))
	}

	casbinEnforcer, err := casbin.NewEnforcer(option.Cfg.AuthConfigPath, adapter)
	if err != nil {
		option.Logger.Fatal("error while creating a new casbin enforcer\n", logger.Error(err))
	}

	casbinEnforcer.GetRoleManager().AddMatchingFunc("keyMatch", util.KeyMatch)
	casbinEnforcer.GetRoleManager().AddMatchingFunc("keyMatch3", util.KeyMatch3)

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	jwtHandler := tokens.JWTHandler{
		SignInKey: option.Cfg.SignInKey,
		Log:       option.Logger,
	}

	handlerV1 := v1.New(&v1.HandlerV1Config{
		InMemoryStorage: option.InMemory,
		Log:             option.Logger,
		ServiceManager:  option.ServiceManager,
		Cfg:             option.Cfg,
		JwtHandler:      jwtHandler,
		Casbin:          casbinEnforcer,
	})

	api := router.Group("/v1")

	api.Use(middleware.Auth(casbinEnforcer, option.Cfg))

	api.POST("/register", handlerV1.Register)                   //unauthorized
	api.GET("/verify/:email/:code", handlerV1.Verify)           //unauthorized
	api.POST("/login", handlerV1.Login)                         //unauthorized
	api.POST("/user/create", handlerV1.CreateUser)              //admin
	api.GET("/user/:id", handlerV1.GetUserById)                 //admin
	api.PUT("/user/update/:id", handlerV1.UpdateUser)           //user
	api.DELETE("/user/delete/:id", handlerV1.DeleteUser)        //user
	api.GET("/users/:page/:limit/:filter", handlerV1.ListUsers) //admin
	api.POST("/user/password/change", handlerV1.ChangePassword) //user

	url := ginSwagger.URL("swagger/doc.json")
	api.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	return router

}
