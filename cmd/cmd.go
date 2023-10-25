package main

import (
	"context"
	"encoding/json"
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
	"time"
)

func main() {
	listenAddr4 := "0.0.0.0:30009" // #nosec G102
	listenAddr3 := "0.0.0.0:30005" // #nosec G102
	listenAddr := "0.0.0.0:30001"  // #nosec G102
	listenAddr2 := "0.0.0.0:30000" // #nosec G102

	listener, err := net.ListenTCP("tcp", getResolvedAddresses(listenAddr))
	listener2, err := net.ListenTCP("tcp", getResolvedAddresses(listenAddr2))
	listener3, err := net.ListenTCP("tcp", getResolvedAddresses(listenAddr3))
	listener4, err := net.ListenTCP("tcp", getResolvedAddresses(listenAddr4))

	if err != nil {
		log.Printf("Failed to listen: %v", err)
	}
	defer func() {
		listener.Close()
		listener2.Close()
		listener3.Close()
		listener4.Close()
	}()

	log.Printf("Listening on %s", listenAddr)
	log.Printf("Listening on %s", listenAddr2)
	log.Printf("Listening on %s", listenAddr3)
	log.Printf("Listening on %s", listenAddr4)

	listeners := []net.TCPListener{*listener, *listener2, *listener3, *listener4}

	for _, lis := range listeners {
		go func(l net.TCPListener) {
			for {
				clientConn, innerError := l.AcceptTCP()
				if innerError != nil {
					log.Printf("Failed to accept client connection: %v", err)
					os.Exit(1)
				}

				go handleClient(clientConn)
			}
		}(lis)
	}

	chExit := make(chan os.Signal, 1)
	signal.Notify(chExit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-chExit:
		log.Println("Example EXITING...Bye.")
		os.Exit(1)
	}

}

func handleClient(clientConn *net.TCPConn) {
	port := strings.Split(clientConn.LocalAddr().String(), ":")[1]

	if port == "30000" {
		if _, err := clientConn.Write([]byte("Health Check Succeed\n")); err != nil {
			log.Printf("Failed to write to client: %v", err)
		}
		return
	}

	var err error
	var client = sidecar.NewClient(context.Background())

	var response *http.Response
	if response, err = client.SendAPIRequest(port); err != nil || response.StatusCode != 200 {
		log.Printf("Failed to get backends endpoints")
	}

	var hostName string
	if err != nil || response == nil || response.StatusCode != 200 {
		hostName = "localhost"
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
			client.StartInstance(responseBody.InstanceID)
		case sidecar.STARTING:
			log.Printf("Instance is starting, waiting for instance to be available")
			if _, err = clientConn.Write([]byte("Instance is starting, waiting for instance to be available\n")); err != nil {
				log.Printf("Failed to write to client: %v", err)
			}
			return
		}

		for _, sc := range responseBody.ServiceComponents {
			if strings.Contains(sc.Alias, "postgres") {
				hostName = sc.NodesEndpoints[0].Endpoint
				break
			}
		}
	}

	hostName = hostName + ":5432"

	var rconn *net.TCPConn

	retryCount := 0
	for retryCount < 22 {
		// connect to remote server
		rconn, err = net.DialTCP("tcp", nil, getResolvedAddresses(hostName))
		if err != nil {
			log.Printf("Remote connection failed: %s", err)

			time.Sleep(15 * time.Second)
			retryCount++
		} else {
			break
		}
	}

	if err != nil {
		log.Printf("Fail to connect remote within timeout: %s", err)
		return
	}

	log.Printf("try connect to %s", hostName)
	// proxying data
	go handleIncomingConnection(clientConn, rconn)
	go handleResponseConnection(rconn, clientConn)

	defer func() {
		if response != nil {
			if closeErr := response.Body.Close(); closeErr != nil {
				log.Printf("Failed to close response body: %v", closeErr)
			}
		}

		// close connections later
	}()
}

func handleIncomingConnection(src, dst *net.TCPConn) {
	// directional copy (64k buffer)
	buff := make([]byte, 0xffff)

	for {
		n, err := src.Read(buff)
		if err != nil {
			log.Printf("Read failed '%s'\n", err)
			return
		}
		b, err := getModifiedBuffer(buff[:n])
		if err != nil {
			log.Printf("%s\n", err)
			err = dst.Close()
			if err != nil {
				log.Printf("connection closed failed '%s'\n", err)
			}
			return
		}

		n, err = dst.Write(b)
		if err != nil {
			log.Printf("Write failed '%s'\n", err)
			return
		}
	}
}

// Proxy.handleResponseConnection
func handleResponseConnection(src, dst *net.TCPConn) {
	// directional copy (64k buffer)
	buff := make([]byte, 0xffff)

	for {
		n, err := src.Read(buff)
		if err != nil {
			log.Printf("Read failed '%s'\n", err)
			return
		}
		b := setResponseBuffer(buff[:n])

		n, err = dst.Write(b)
		if err != nil {
			log.Printf("Write failed '%s'\n", err)
			return
		}
	}
}

func getModifiedBuffer(buffer []byte) (b []byte, err error) {
	return buffer, nil
}

func setResponseBuffer(buffer []byte) (b []byte) {
	if len(buffer) > 0 && string(buffer[0]) == "Q" {
		return nil
	}

	return buffer
}

// ResolvedAddresses of host.
func getResolvedAddresses(host string) *net.TCPAddr {
	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		log.Printf("ResolveTCPAddr of host:", err)
	}
	return addr
}
