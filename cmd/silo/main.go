package main

import (
	"flag"
	"fmt"
	"github.com/stockyard-dev/stockyard-silo/internal/server"
	"github.com/stockyard-dev/stockyard-silo/internal/store"
	"log"
	"net/http"
	"os"
)

func main() {
	portFlag := flag.String("port", "", "")
	dataFlag := flag.String("data", "", "")
	flag.Parse()
	port := os.Getenv("PORT")
	if port == "" {
		port = "9700"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./silo-data"
	}
	if *dataFlag != "" {
		dataDir = *dataFlag
	}
	if *portFlag != "" {
		port = *portFlag
	}
	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("silo: %v", err)
	}
	defer db.Close()
	srv := server.New(db, server.DefaultLimits(), dataDir)
	fmt.Printf("\n  Silo — Self-hosted data isolation and partitioning layer\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Questions? hello@stockyard.dev — I read every message\n\n", port, port)
	log.Printf("silo: listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
