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
	"strings"
	"syscall"
)

/**var (
	count int64 = 0
)**/

func main() {
	listenAddr3 := "0.0.0.0:30009" // #nosec G102
	listenAddr2 := "0.0.0.0:30005" // #nosec G102
	listenAddr1 := "0.0.0.0:30001" // #nosec G102
	listenAddr := "0.0.0.0:30000"  // #nosec G102

	//connStr2 := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/omnistratemetadatadb", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"))

	listener3, err := net.Listen("tcp", listenAddr3)
	listener2, err := net.Listen("tcp", listenAddr2)
	listener1, err := net.Listen("tcp", listenAddr1)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Printf("Failed to listen: %v", err)
	}
	defer func() {
		listener3.Close()
		listener2.Close()
		listener1.Close()
		listener.Close()
	}()

	log.Printf("Listening on %s", listenAddr3)
	log.Printf("Listening on %s", listenAddr2)
	log.Printf("Listening on %s", listenAddr1)
	log.Printf("Listening on %s", listenAddr)

	for {
		var innerError error
		clientConn1, innerError := listener1.Accept()
		if innerError != nil {
			log.Printf("Failed to accept client connection: %v", innerError)
			os.Exit(1)
		}

		clientConn2, innerError := listener1.Accept()
		if innerError != nil {
			log.Printf("Failed to accept client connection: %v", innerError)
			os.Exit(1)
		}

		clientConn3, innerError := listener1.Accept()
		if innerError != nil {
			log.Printf("Failed to accept client connection: %v", innerError)
			os.Exit(1)
		}

		clientConn, innerError := listener.Accept()
		if innerError != nil {
			log.Printf("Failed to accept client connection: %v", innerError)
			os.Exit(1)
		}

		go func() {
			if _, err = clientConn.Write([]byte("Health Check Succeed\n")); err != nil {
				log.Printf("Failed to write to client: %v", err)
			}

			defer clientConn.Close()
		}()

		go handleClient(clientConn1)
		go handleClient(clientConn2)
		go handleClient(clientConn3)

	}

	chExit := make(chan os.Signal, 1)
	signal.Notify(chExit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-chExit:
		log.Println("Example EXITING...Bye.")
		os.Exit(1)
	}
}

// func handleClient(clientConn net.Conn, connStr string, connStr2 string) {
func handleClient(clientConn net.Conn) {

	port := strings.Split(clientConn.LocalAddr().String(), ":")[1]
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	reqLen, err := clientConn.Read(buf)

	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	reqLen = reqLen
	fmt.Printf("Received data: %v\n", string(buf[:reqLen]))

	var client = sidecar.NewClient(context.Background())

	var response *http.Response
	if response, err = client.SendAPIRequest(port); err != nil || response.StatusCode != 200 {
		log.Printf("Failed to get backends endpoints")
	}

	var connStr string // Connection string to the database
	if response == nil || response.StatusCode != 200 {
		fmt.Sprintf("host=%s port=5432 user=%s dbname=postgres sslmode=disable password=%s",
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
			log.Printf("Instance is paused, waking up instance " + responseBody.InstanceID)

			if _, err = clientConn.Write([]byte("Instance is paused, waking up instance\n")); err != nil {
				log.Printf("Failed to write to client: %v", err)
			}
			return
		case sidecar.STARTING:
			log.Printf("Instance is starting, waiting for instance to be available")
			if _, err = clientConn.Write([]byte("Instance is starting, waiting for instance to be available\n")); err != nil {
				log.Printf("Failed to write to client: %v", err)
			}
			return
		}

		var hostName string
		for _, sc := range responseBody.ServiceComponents {
			if strings.Contains(sc.Alias, "postgres") {
				hostName = sc.NodesEndpoints[0].Endpoint
				break
			}
		}

		connStr = fmt.Sprintf("host=%s port=5432 user=%s dbname=postgres sslmode=disable password=%s",
			hostName, os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"))
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

	select {
	case <-done:
		log.Printf("Traffic forwarding completed to %s", connStr)
		if _, err = clientConn.Write([]byte("Connected to backend\n")); err != nil {
			log.Printf("Failed to write to client: %v", err)
		}
		return
	}
}
