package router

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/vizurth/concurency_calc/internal/server/handler"
	"github.com/vizurth/concurency_calc/internal/worker"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Host string
	Port string
}

type Router struct {
	config Config
	Router *mux.Router
	client *http.Client
}

func NewRouter(config Config) *Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/calculate", handler.CalculateHandler).Methods("POST")               //Done
	r.HandleFunc("/api/v1/expressions/{id}", handler.GetExpressionByIdHandler).Methods("GET") //Done
	r.HandleFunc("/api/v1/expressions", handler.GetExpressionsHandler).Methods("GET")         //Done
	r.HandleFunc("/internal/task", handler.GetTaskHandler).Methods("GET")                     //Done
	r.HandleFunc("/internal/task", handler.UpdateTaskHandler).Methods("PUT")                  //Done

	client := &http.Client{
		Timeout: 10 * time.Second, // Устанавливаем таймаут для запросов
	}

	return &Router{config: config, Router: r, client: client}
}

func (router *Router) Start() {
	log.Println("Starting server...")

	workers := os.Getenv("COMPUTING_POWER")

	if workers == "" {
		workers = "3"
	}

	workerInt, err := strconv.Atoi(workers)

	if err != nil {
		fmt.Errorf("Something went with Worker in main()")
	}
	for i := 0; i < workerInt; i++ {
		go worker.Worker(i+1, router.config.Host, router.config.Port, router.client)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", router.config.Host, router.config.Port), router.Router))
}
