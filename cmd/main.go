package main

import (
	"fmt"
	rds "github.com/gomodule/redigo/redis"
	"myproject/api-gateway/api"
	"myproject/api-gateway/config"
	"myproject/api-gateway/pkg/etc"
	"myproject/api-gateway/pkg/logger"
	"myproject/api-gateway/services"
	"myproject/api-gateway/storage/redis"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel, "api-gateway")

	serviceManager, err := services.NewServiceManager(cfg)

	if err != nil {
		log.Error("gRPC dial error", logger.Error(err))
	}

	redisPool := rds.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (rds.Conn, error) {
			c, err := rds.Dial("tcp", fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort))
			if err != nil {
				panic(err)
			}
			return c, nil
		},
	}

	//db, _, err := db2.ConnectToDB(cfg)
	//if err != nil {
	//	log.Fatal("cannot connect to DB", logger.Error(err))
	//	panic(err)
	//}
	fmt.Println(etc.GenerateHashPassword("string500"))

	server := api.New(api.Option{
		InMemory:       redis.NewRedisRepo(&redisPool),
		Cfg:            cfg,
		Logger:         log,
		ServiceManager: serviceManager,
	})

	if err := server.Run(cfg.HTTPPort); err != nil {
		log.Fatal("cannot run http server", logger.Error(err))
		panic(err)
	}

}
