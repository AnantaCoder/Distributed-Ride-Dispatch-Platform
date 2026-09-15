//1. load config
//2. initialize the infra connections
//3. dependancy injections
//4. setup the router
//5. asyncronous server  starter
//6. shutdown

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/config"
	"github.com/AnantaCoder/Distributed-Ride-Dispatch-Platform/internal/driver"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)



func main(){
	 
	//loading config
	config := config.LoadConfig()
	log.Println("config loaded : ", config)

	ctx := context.Background() // infra initialize
	//database
	dbURL := config.Postgres.DSN()
	pgPool, err := pgxpool.New(ctx , dbURL) // connecting to databse using connection pool 
	if err != nil {
		log.Fatal("failed to connect to postgres database : ", err)
	}
	defer pgPool.Close()

	//redis
	redisClient := redis.NewClient(
		&redis.Options{
			Addr: config.Redis.Addr(),
			Password: config.Redis.Password,
			DB: config.Redis.DB,
		})

	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("failed to connect to redis database : ", err)
	}
	log.Println("redis connected successfully")


	//deps injection
	repository := driver.NewRepository(pgPool)
	service := driver.NewService(repository, redisClient)
	handler := driver.NewHandler(service)

	// routers 
	router := chi.NewRouter()

	// middle wares 
	router.Use(middleware.RequestID)	
	router.Use(middleware.Logger)	
	router.Use(middleware.Recoverer)	

	handler.RegisterRoutes(router)


	// async starter 
	port := config.Services.DriverServicePort
	log.Println("starting driver service on port : ", port)
	address := fmt.Sprintf(":%d", port)
	log.Println("server address : ", address)

	server := &http.Server{
		Addr:              address,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
	}
// separate gorutine to start the server 
	go func ()  {
		log.Println("driver service running on ")
		if err := server.ListenAndServe(); err != nil {
			log.Fatal("failed to start driver service : ", err)
		}
	}()

	//shutdown 

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("server shutdown failed : ", err)
	}
	log.Println("server exited successfully ")

	
	



	

	

	
	

	



}