package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"
	"gitlab.com/dipankardas011/universalci-cd/internal/api"
	"gitlab.com/dipankardas011/universalci-cd/internal/ui"
)

func main() {

	http.HandleFunc("/api/build", api.MakeHTTPHandler(api.BuildNow))
	http.HandleFunc("/api/create", api.MakeHTTPHandler(api.CreatePipeline))
	http.HandleFunc("/api/delete", api.MakeHTTPHandler(api.DeletePipeline))
	http.HandleFunc("/api/list", api.MakeHTTPHandler(api.GetAllPipelines))
	http.HandleFunc("/api/get", api.MakeHTTPHandler(api.GetBuildDetails))
	http.HandleFunc("/api/docs", api.MakeHTTPHandler(api.Docs))

	http.HandleFunc("/ui", api.MakeHTTPHandler(ui.HomeUI))
	http.HandleFunc("/ui/list", api.MakeHTTPHandler(ui.ListUI))
	http.HandleFunc("/ui/create", api.MakeHTTPHandler(ui.CreateUI))
	http.HandleFunc("/ui/delete", api.MakeHTTPHandler(ui.DeleteUI))

	http.HandleFunc("/auth", api.MakeHTTPHandler(nil)) // TODO: need to add auth handler

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},                             // Allow all origins
		AllowedMethods: []string{"GET", "POST", "PUT", "OPTIONS"}, // Allow GET, OPTIONS methods
	})

	s := &http.Server{
		Addr:           ":8080",
		MaxHeaderBytes: 1 << 20,
		Handler:        c.Handler(http.DefaultServeMux),
	}

	log.Println("universal CI-CD server listening on port 8080")

	log.Printf("Started to serve on port {%v}\n", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}
