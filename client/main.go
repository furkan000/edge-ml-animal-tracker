package main

import (
	"log"
	"net"
	"os"
	"strings"
)

/**
 * This code is used for the small software on the edge, that is tasked with send all the
 */

func main() {
	cameraNodeTarget := os.Getenv("CLIENT_HOG_CAMERA_IP") // Default: "localhost:3333"
	fogNodeTarget := os.Getenv("CLIENT_HOG_SERVER_IP")    // Default: "localhost:3444"
	if cameraNodeTarget == "" || fogNodeTarget == "" {
		log.Fatalln("Environmental variables not initialized.")
	}

	// start handler
	handler := startAndRunNewTxHandler(fogNodeTarget)

	// Start listener
	pc, err := net.ListenPacket(
		"udp",
		cameraNodeTarget,
	)

	if err != nil {
		log.Printf("Error starting server: %v\n", err.Error())
		os.Exit(1)
	}

	defer func(l net.PacketConn) {
		err := l.Close()
		if err != nil {
			log.Println("Error closing server.")
		}
	}(pc)

	log.Printf("Server started successfully: %s", cameraNodeTarget)

	// TODO: a future-proof approach would set them functionally upon initialization from a list by either asking the\
	//   cloud or using a local config file
	approvedCameras := []string{
		"localhost",
		"127.0.0.1",
	}

	for {
		p := make([]byte, 1500)

		n, addr, err := pc.ReadFrom(p)
		if err != nil {
			log.Println("Error with new request: ", err.Error())
			os.Exit(1)
		}

		// check if addr is in the approved list of hosts
		approved := false
		for _, val := range approvedCameras {
			if strings.Contains(addr.String(), val) {
				approved = true
			}
		}

		if !approved {
			log.Println("WARNING: Unapproved user tried to connect!")
			continue
		}

		log.Printf("New message: %s\n", string(p[:n]))
		go handler.handleRequest(p[:n])
	}
}
