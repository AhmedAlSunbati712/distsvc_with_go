package main

import (
	"log"
	"github.com/ahmedalsunbati712/proglog/internal/server"
)

func main() {
	srv := server.NewHTTPServer(":8080")
	log.Fatal(srv.ListenAndServe())
}