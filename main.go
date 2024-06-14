package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"project/controllers"
	"project/initializers"
	"project/routes"
	services "project/service"

	"project/grpcApi"

	"project/pb"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/grpc"
)

var (
	Server      *gin.Engine
	MongoClient *mongo.Client
	RedisClient *redis.Client
	ctx         context.Context

	AuthCollection      *mongo.Collection
	AuthController      controllers.AuthController
	AuthRouteController routes.AuthRouteController
	AuthServiceImpl     services.AuthServiceImpl

	UserController      controllers.UserController
	UserRouteController routes.UserRouteController
	UserServiceImpl     services.UserServiceImpl

	PostCollection      *mongo.Collection
	PostController      controllers.PostController
	PostRouteController routes.PostRouteController
	PostServices        services.PostServices
	PostServiceImpl     services.PostServiceImpl
)

func init() {
	Server = gin.Default()
	ctx = context.TODO()
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatal("Could not load environment variables", err)
	}

	//connect Mongo
	MongoClient, err = initializers.ConnectMongo(config, ctx)
	if err != nil {
		log.Fatal("Failed to connect to mongo")
	}
	if err := MongoClient.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}
	fmt.Println("MongoDB successfully connected...")

	//connect Redis
	RedisClient = initializers.ConnectRedis(config)
	if _, err := RedisClient.Ping(ctx).Result(); err != nil {
		panic(err)
	}
	err = RedisClient.Set(ctx, "test", "Welcome to Golang with Redis and MongoDB", 0).Err()
	if err != nil {
		panic(err)

	}
	fmt.Println("Redis client connected successfully...")

	AuthCollection = MongoClient.Database("golang_mongodb").Collection("users")
	PostCollection = MongoClient.Database("golang_mongodb").Collection("posts")

	UserServiceImpl = services.NewUserServiceImpl(AuthCollection, ctx)
	UserController = controllers.NewUserController(&UserServiceImpl)

	AuthServiceImpl = services.NewAuthServiceImpl(AuthCollection, ctx)
	AuthController = controllers.NewAuthController(&AuthServiceImpl, &UserServiceImpl, ctx, AuthCollection)
	AuthRouteController = routes.NewAuthRouteController(AuthController)

	PostServiceImpl = services.NewPostServiceImpl(PostCollection, ctx)
	PostController = controllers.NewPostController(&PostServiceImpl)
	PostRouteController = routes.NewPostRouteController(PostController)

}

func main() {
	config, err := initializers.LoadConfig(".")

	if err != nil {
		log.Fatal("Could not load config", err)
	}

	defer MongoClient.Disconnect(ctx)

	go StartGrpcServer(config)
	StartGinServer(config)
}

func StartGinServer(config initializers.Config) {
	value, err := RedisClient.Get(ctx, "test").Result()

	if err == redis.Nil {
		fmt.Println("key: test does not exist")
	} else if err != nil {
		panic(err)
	}

	router := Server.Group("/api")
	router.GET("/healthchecker", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": value})
	})

	AuthRouteController.AuthRoute(router, &UserServiceImpl)
	UserRouteController.UserRoute(router, &UserServiceImpl)
	PostRouteController.PostRoute(router, &UserServiceImpl)
	log.Fatal(Server.Run(":" + config.Port))
}

func StartGrpcServer(config initializers.Config) {
	authServer := grpcApi.NewGrpcAuthServer(&AuthServiceImpl, &UserServiceImpl)
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Error in listening: %v \n", err)
	}

	go func() {
		fmt.Println("Starting gRPC server on localhost:8080")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	gwMux := runtime.NewServeMux()
	err = pb.RegisterAuthServiceHandlerServer(ctx, gwMux, authServer)
	if err != nil {
		log.Fatalf("Failed to register gRPC Gateway: %v", err)
	}

	fmt.Println("Starting gRPC Gateway on localhost:9090")
	if err := http.ListenAndServe("localhost:9090", gwMux); err != nil {
		log.Fatalf("Failed to start gRPC Gateway: %v", err)
	}
}
