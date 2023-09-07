package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/omnistrate/pg-proxy/pkg/sidecar"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

/**var (
	count int64 = 0
)**/

func main() {
	listenAddr := "0.0.0.0:3001" // #nosec G102

	//connStr2 := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/omnistratemetadatadb", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"))

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Printf("Failed to listen: %v", err)
	}
	defer listener.Close()

	log.Printf("Listening on %s", listenAddr)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept client connection: %v", err)
			continue
		}

		//go handleClient(clientConn, connStr, connStr2)
		go handleClient(clientConn)
	}

	chExit := make(chan os.Signal, 1)
	signal.Notify(chExit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-chExit:
		log.Println("Example EXITING...Bye.")
	}
}

// func handleClient(clientConn net.Conn, connStr string, connStr2 string) {
func handleClient(clientConn net.Conn) {
	var client = sidecar.NewClient(context.Background())

	var response *http.Response
	var err error
	if response, err = client.SendAPIRequest(); err != nil || response.StatusCode != 200 {
		log.Printf("Failed to get backends endpoints")
	}

	var connStr string // Connection string to the database
	if response == nil || response.StatusCode != 200 {
		fmt.Sprintf("host=%s port=5432 user=%s dbname=omnistratemetadatadb sslmode=disable password=%s",
			"nohost.com", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"))
	} else {

		var body []byte
		if body, err = io.ReadAll(response.Body); err != nil {
			log.Printf("Failed to read response body")
		}

		responseBody := &sidecar.InstanceStatus{}

		if err = json.Unmarshal(body, &responseBody); err != nil {
			log.Printf("Failed to unmarshal response body")
		}

		log.Print(responseBody)

		switch responseBody.Status {
		case sidecar.PAUSED:
			log.Printf("Instance is paused, waking up instance")
			return
		case sidecar.STARTING:
			log.Printf("Instance is starting, waiting for instance to start")
			return
		}

		connStr = fmt.Sprintf("host=%s port=5432 user=%s dbname=omnistratemetadatadb sslmode=disable password=%s",
			responseBody.ServiceComponents[0].NodesEndpoints[0].Endpoint, os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"))
	}

	defer func() {
		if response != nil {
			if closeErr := response.Body.Close(); closeErr != nil {
				log.Printf("Failed to close response body: %v", closeErr)
			}
		}

		clientConn.Close()
	}()

	var db *sql.DB

	//if count%2 == 0 {
	db, err = sql.Open("postgres", connStr)
	log.Printf("Connecting to %s", connStr)
	//} else {
	//	db, err = sql.Open("mysql", connStr2)
	//	log.Printf("Connecting to %s", connStr2)

	//}

	if err != nil {
		log.Printf("Failed to connect to the database: %v", err)
		return
	}
	//count++
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
