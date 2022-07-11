package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/tcp"
	"log"
	"strings"
	"time"
)

const maxBufferLength = 10

// TxHandler is a Transmission Handler
type TxHandler struct {
	newChan        chan []byte // communication channel that lead to the
	currentFogNode string
	deviceUUID	   int
	sliceTrackedNames []string
}

func startAndRunNewTxHandler(target string, deviceUUID int, initialTrackedNames []string) TxHandler {
	nh := TxHandler{
		newChan:        make(chan []byte, 25),
		currentFogNode: "",
		deviceUUID: deviceUUID,
		sliceTrackedNames: initialTrackedNames,
	}

	go nh.startHandler(target)

	return nh
}

func (h *TxHandler) startHandler(target string) {
	buffer := make([][]byte, 0, maxBufferLength+1)

	// Timeout waiting time of 50ms, 0.25s, 0.5s, 1s, 1.5s, 4s, 7s
	timeoutArray := [7]time.Duration{50, 250, 500, 1000, 1500, 4000, 7000}
	failures := 0

	for {
		// Get the next message to handle
		if len(buffer) == 0 {
			// No messages in the buffer. Waiting for the next one.
			element := <-h.newChan

			// add element to buffer
			buffer = append(buffer, element)

		} else {
			// We do have elements in the buffer. Accept the new message only if the channel is not empty
			for {
				if len(h.newChan) == 0 {
					break
				}
				element := <-h.newChan
				// add element to buffer (take maximum buffer size into account)
				buffer = append(buffer, element)

				// if buffer size is too large, drop the first element
				if len(buffer) == maxBufferLength+1 {
					log.Println("WARNING: Buffer full; dropping oldest element.")
					_, buffer = buffer[0], buffer[1:]
				}
			}
		}

		// get the address of the Fog Node (example localhost:5335)
		h.currentFogNode = target

		// Attempt transmission
		err := h.sendData(buffer[0])

		// If transmission failed, block transmission
		if err != nil {
			log.Println("Transmission failed: Unable to connect with target. Must retry later.")

			// increasing timeouts to avoid DDoS
			time.Sleep(timeoutArray[failures] * time.Millisecond)
			if failures < len(timeoutArray)-1 {
				failures += 1
			}
		} else {
			failures = 0
			if len(buffer) > 1 {
				_, buffer = buffer[0], buffer[1:]
			} else {
				buffer = buffer[:0]
			}
		}
	}
}

func (h TxHandler) sendData(payload []byte) error {
	path := "/a"

	co, err := tcp.Dial(h.currentFogNode)
	if err != nil {
		log.Printf("Error dialing: %v\n", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := co.Post(ctx, path, message.AppJSON, bytes.NewReader(payload))
	if err != nil {
		log.Printf("Error sending request: %v", err)
		// start a go routine that decreases in frequency and checks if server is alive
		// if it is, the routine sends the data
		return err
	}
	log.Printf("Response payload: %v", resp.String())
	return nil
}


type Animal struct {
	DetectionID   int    `json:"detection_id"`
	DeviceUuid    int    `json:"device_uuid"`
	DetectionTime string `json:"detection_time"`
	DetectedAnimal string  `json:"detected_object"`
	Temperature    float64 `json:"temperature"`
}


func (h TxHandler) handleRequest(buf []byte) {
	// Parse element from json object
	var animal Animal

	err := json.Unmarshal(buf, &animal)
	if err != nil {
		log.Fatalln(err)
	}

	// add uuid to buffer
	animal.DeviceUuid = h.deviceUUID

	// check if the detected animal is in the approved list, then put it in the channel
	for _, val := range h.sliceTrackedNames {
		if strings.Contains(animal.DetectedAnimal, val) {
			// put buffer into channel
			h.newChan <- buf
			return
		}
	}


}
