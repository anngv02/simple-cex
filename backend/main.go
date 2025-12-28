package main

import (
	"log"
	"simple-cex/api"    // Import package api
	"simple-cex/engine" // Import package engine
)


func main() {
	// 1. Kết nối DB
	if err := InitDB(); err != nil {
		log.Fatal("Cannot connect to DB:", err)
	}
	defer db.Close()
	log.Println("DB connected")

	// 2. Khởi tạo Engine (Core Logic)
	tradeEngine := engine.NewEngine(db)
	
	// 3. Khởi tạo API Server (Lớp giao tiếp)
	server := api.NewServer(tradeEngine)

	// 4. Chạy Server tại port 8010
	log.Println("Starting server on 0.0.0.0:8010")
	if err := server.Start("0.0.0.0:8010"); err != nil {
		log.Fatal("Cannot start server:", err)
	}
}