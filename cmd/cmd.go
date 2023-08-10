package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	listenAddr := "0.0.0.0:3001" // #nosec G102
	connStr := fmt.Sprintf("host=localhost port=5432 user=%s dbname=omnistratemetadatadb sslmode=disable password=%s",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD")) // Connection string to the database

	log.Printf("Connecting to %s", connStr)

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer listener.Close()

	log.Printf("Listening on %s", listenAddr)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept client connection: %v", err)
			continue
		}

		go handleClient(clientConn, connStr)
	}

	chExit := make(chan os.Signal, 1)
	signal.Notify(chExit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-chExit:
		log.Println("Example EXITING...Bye.")
	}
}

func handleClient(clientConn net.Conn, connStr string) {
	defer clientConn.Close()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("Failed to connect to the database: %v", err)
		return
	}
	defer db.Close()

	dbConn, err := db.Conn(context.Background())
	if err != nil {
		log.Printf("Failed to create DB connection: %v", err)
		return
	}
	defer dbConn.Close()

	done := make(chan struct{})
	go func() {
		//Using select 1 to mimic data parse and transferring
		_, err := dbConn.PrepareContext(context.Background(), "SELECT 1")
		if err != nil {
			log.Printf("Failed to copy from client to database: %v", err)
		}
		done <- struct{}{}
	}()

	_, err = clientConn.Write([]byte("Ready to forward traffic\n"))
	if err != nil {
		log.Printf("Failed to write to client: %v", err)
		return
	}

	select {
	case <-done:
		log.Printf("Traffic forwarding completed to %s", connStr)
		return
	}
}
