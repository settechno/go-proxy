package main

import (
	"fmt"
	"log"
	"net"
	"proxy/internal/application"
)

var (
	app *application.App
)

func main() {
	app = application.NewApp()

	// Запускаем SOCKS5-сервер
	address := fmt.Sprintf(":%d", app.Config.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Println("SOCKS5 server is running on " + address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go func() {
			err := application.HandleRequest(conn, app)
			if err != nil {
				log.Printf("Request handling failed: %v", err)
			}
		}()
	}
}
