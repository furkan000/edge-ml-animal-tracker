package main

import (
	"bytes"
	"encoding/json"
	"github.com/plgd-dev/go-coap/v2"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

/**
 * This code is used for the small software on the edge, that is tasked with send all the
 */



func loggingMiddleware(next mux.Handler) mux.Handler {
	return mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		log.Printf("ClientAddress %v, %v\n", w.Client().RemoteAddr(), r.String())
		next.ServeCOAP(w, r)
	})
}

func handleA(w mux.ResponseWriter, r *mux.Message) {
	var animal Animal
	err := json.NewDecoder(r.Body).Decode(&animal)
	if err != nil {
		log.Printf("cannot decode json object: %v", err)
		return
	}

	// TODO: update list

	err = w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("gg fam")))
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}



func main() {
	cameraNodeTarget := os.Getenv("CLIENT_HOG_CAMERA_IP") // Default: "localhost:3333"
	fogNodeTarget := os.Getenv("CLIENT_HOG_SERVER_IP")    // Default: "localhost:3444"
	deviceUUIDString := os.Getenv("CLIENT_HOG_DEVICE_UUID")    // Default: "352" (Doesn't matter)
	initialTrackedAnimalsString := os.Getenv("CLIENT_HOG_TRACKED_ANIMALS") // Default: "[\"Bear\",\"Racoon\",\"Gazelle\"]"
	clientHogConfigRxPort := os.Getenv("CLIENT_HOG_LOCAL_CONFIG_RECEIVER_PORT") // Default: ":3555"
	if cameraNodeTarget == "" || fogNodeTarget == "" ||  deviceUUIDString == "" {
		log.Fatalln("Environmental variables not initialized.")
	}

	deviceUUID, err := strconv.Atoi(deviceUUIDString)

	if err != nil {
		log.Fatalf("deviceUUID invalid: %v", err)
	}

	var trackedAnimalsList []string

	err = json.Unmarshal([]byte(initialTrackedAnimalsString), &trackedAnimalsList)
	if err != nil {
		return
	}

	// start handler
	handler := startAndRunNewTxHandler(fogNodeTarget, deviceUUID, trackedAnimalsList)

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

	// add listener, that is responsible for setting the list of tracked animals
	go func() {
		r := mux.NewRouter()
		r.Use(loggingMiddleware)
		err = r.Handle("/a", mux.HandlerFunc(handleA))
		if err != nil {
			return
		}

		log.Fatal(coap.ListenAndServe("tcp", clientHogConfigRxPort, r))
	}()

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
