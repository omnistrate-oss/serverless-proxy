package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/omnistrate-oss/serverless-proxy/internal/sidecar"
	"github.com/pkg/errors"
)

/**
 * In high level, the proxy does following steps:
 * 1. Start frontend(end client to proxy) TCP listeners.
 * 2. Discover backend instance's endpoint via mapped proxy port.
 *   2.a If backend instance is paused, starting the backend instance and holding frontend connections until backend instance is active.
 * 3. Start backend(proxy to serverless resource instance) TCP channel.
 * 4. Forward data from frontend to backend and forward response data from backend to frontend.
 */
func main() {

	// Step 1: Start frontend TCP listener from port 30000,
	// note that 30000 is reserved for Omnistrate health check and will not be assigned to any backend instances,
	// and you can leverage this port for internal use case.
	listeners := []net.TCPListener{}
	for i := 0; i <= 9; i++ {
		listenAddr := "0.0.0.0:3000" + strconv.Itoa(i) // #nosec G102

		// Setup frontend TCP listener
		listener, err := net.ListenTCP("tcp", getResolvedAddresses(listenAddr))
		if err != nil {
			log.Printf("Failed to listen: %v", err)
		}
		log.Printf("Listening on %s", listenAddr)

		listeners = append(listeners, *listener)
	}

	// Initialize Omnistrate sidecar sidecarClient
	var sidecarClient = sidecar.NewClient(context.Background())

	for _, lis := range listeners {
		go func(l net.TCPListener) {
			for {
				frontEndConnection, innerError := l.AcceptTCP()
				if innerError != nil {
					log.Printf("Failed to accept front end connection: %v", innerError)
					os.Exit(1)
				}

				go handleClient(frontEndConnection, sidecarClient)
			}
		}(lis)
	}

	chExit := make(chan os.Signal, 1)
	signal.Notify(chExit, syscall.SIGINT, syscall.SIGTERM)
	<-chExit
	log.Println("EXITING...Bye.")
	for _, listener := range listeners {
		listener.Close()
	}
	os.Exit(1)

}

func handleClient(frontEndConnection *net.TCPConn, sidecarClient *sidecar.Client) {
	port := strings.Split(frontEndConnection.LocalAddr().String(), ":")[1]

	// Port 30000 is reserved for health check and internal use case
	if port == "30000" {
		if _, err := frontEndConnection.Write([]byte("Health Check Succeed\n")); err != nil {
			log.Printf("Failed to write to client: %v", err)
		}
		return
	}

	err := frontEndConnection.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		log.Printf("Failed to set read deadline: %v", err)
		return
	}

	inputBuffer := make([]byte, 0xffff)
	size, err := frontEndConnection.Read(inputBuffer)

	if err == nil {
		log.Printf("Not mysql connections, returning")
		_ = frontEndConnection.Close()
		return
	}
	if err != nil {
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			log.Printf("Failed to read from client: %v", err)
			return
		}
	}

	err = frontEndConnection.SetReadDeadline(time.Time{})
	if err != nil {
		log.Printf("Failed to set read deadline: %v", err)
		return
	}

	_, err = getModifiedBuffer(inputBuffer[:size])
	if err != nil {
		log.Printf("%s\n", err)
		return
	}

	var serverlessTargetPort string
	var hostName string
	var backendConnection *net.TCPConn
	if os.Getenv("DRY_RUN") == "true" {
		var err error
		hostName = "127.0.0.1"
		serverlessTargetPort = "3306"
		hostName = hostName + ":" + serverlessTargetPort
		backendConnection, err = net.DialTCP("tcp", nil, getResolvedAddresses(hostName))
		if err != nil {
			log.Printf("Remote connection failed: %s", err)
			return
		}
	} else {
		retryCount := 0
		for retryCount < 22 {
			// Step 2: Discover backend instance's endpoint via mapped proxy port.
			var err error
			var response *http.Response
			if response, err = sidecarClient.QueryBackendInstanceStatus(port); err != nil || response.StatusCode != 200 {
				log.Printf("Failed to get backends endpoints")
				return
			}

			var body []byte
			if body, err = io.ReadAll(response.Body); err != nil {
				log.Printf("Failed to read response body")
				return
			}

			responseBody := &sidecar.InstanceStatus{}

			if err = json.Unmarshal(body, &responseBody); err != nil {
				log.Printf("Failed to unmarshal response body")
			}

			log.Printf("Instance response: %s", responseBody)

			switch responseBody.Status {
			// Step 2a: if backend instance is paused, starting the backend instance and holding frontend connections until backend instance is active.
			// In this example, we are using 22 retries with 15 seconds interval to check backend instance status.
			case sidecar.PAUSED:
				log.Printf("Instance is paused, waking up instance")
				resp, err := sidecarClient.StartInstance(responseBody.InstanceID)
				if err != nil {
					log.Printf("Failed to start instance: %v", err)
					return
				}
				defer resp.Body.Close()
			case sidecar.ACTIVE:
				fallthrough
			case sidecar.STARTING:
				fallthrough
			// If status unknown, still try to connect to avoid system glitch.
			case sidecar.UNKNOWN:
				serverlessTargetPort = os.Getenv("TARGET_PORT")
				if serverlessTargetPort == "" {
					log.Printf("Failed to get serverless target port")
					return
				}

				serverlessResourceKey := os.Getenv("SERVERLESS_RESOURCE_KEY")
				if serverlessResourceKey == "" {
					log.Printf("Failed to get serverless resource key")
					return
				}

				log.Printf("Instance is %s, trying to dial TCP.", responseBody.Status)

				for _, sc := range responseBody.ServiceComponents {
					if strings.Contains(sc.Alias, serverlessResourceKey) {
						hostName = serverlessResourceKey + "." + responseBody.InstanceID
						hostName = hostName + ":" + serverlessTargetPort
						// Step 3: connect to backend serverless resource server
						backendConnection, err = net.DialTCP("tcp", nil, getResolvedAddresses(hostName))
						if err != nil {
							log.Printf("Remote connection failed: %s", err)
						}
						break
					}
				}
			default:
				log.Printf("Instance is not in expected status %s, exiting...", responseBody.Status)
				return
			}

			if responseBody.Status == sidecar.ACTIVE {
				break
			}

			if responseBody.Status != sidecar.STARTING && responseBody.Status != sidecar.PAUSED {
				break
			}

			if responseBody.Status == sidecar.STARTING && backendConnection != nil {
				break
			}

			time.Sleep(5 * time.Second)
			retryCount++

			defer func() {
				if response != nil {
					if closeErr := response.Body.Close(); closeErr != nil {
						log.Printf("Failed to close response body: %v", closeErr)
					}
				}
			}()
		}
	}

	if backendConnection == nil {
		log.Printf("Didn't get backend connection established in time, exiting...")
		return
	}

	// Step 4: Forward data from frontend to backend and forward response data from backend to frontend.
	go handleIncomingConnection(frontEndConnection, backendConnection)
	go handleResponseConnection(backendConnection, frontEndConnection)
}

/**
 * This function is used to forward data from frontend to backend. srcChannel is frontend connection, dstChannel is backend connection.
 */
func handleIncomingConnection(srcChannel, dstChannel *net.TCPConn) {
	buff := make([]byte, 0xffff)

	for {
		var b []byte
		n, err := srcChannel.Read(buff)
		if err != nil {
			if err == io.EOF {
				err = dstChannel.Close()
				if err != nil {
					log.Printf("backend connection closed failed '%s'\n", err)
				}
			}
			log.Printf("Read failed '%s'\n", err)
			return
		}

		// Note that you can add any custom logic, like authentication, authorization
		// before sending data to the backend serverless resource server.
		b, err = getModifiedBuffer(buff[:n])
		if err != nil {
			log.Printf("%s\n", err)
			err = dstChannel.Close()
			if err != nil {
				log.Printf("connection closed failed '%s'\n", err)
			}
			return
		}

		_, err = dstChannel.Write(b)
		if err != nil {
			log.Printf("Write failed '%s'\n", err)
			return
		}
	}
}

/**
 * This function is used to forward data from backend to frontend. srcChannel is backend connection, dstChannel is frontend connection.
 */
func handleResponseConnection(srcChannel, dstChannel *net.TCPConn) {
	// directional copy (64k buffer)
	buff := make([]byte, 0xffff)

	for {
		n, err := srcChannel.Read(buff)
		if err != nil {
			log.Printf("Read failed '%s'\n", err)
			return
		}
		b := setResponseBuffer(buff[:n])

		_, err = dstChannel.Write(b)
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
		log.Printf("ResolveTCPAddr of host:%s: %v", host, err)
	}
	return addr
}
