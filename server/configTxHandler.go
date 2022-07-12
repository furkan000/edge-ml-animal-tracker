package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/tcp"
	"log"
	"time"
)

const maxBufferLength = 10

// TxHandler is a Transmission Handler
type TxHandler struct {
	newChan  chan []byte // communication channel that lead to the
	clientIP string
}

func startAndRunNewTxHandler(target string) TxHandler {
	nh := TxHandler{
		newChan:  make(chan []byte, 25),
		clientIP: target,
	}

	go nh.startHandler()

	return nh
}

func (h *TxHandler) startHandler() {
	buffer := make([][]byte, 0, maxBufferLength+1)

	// Timeout waiting time of 1s, 1.5s, 4s, 7s, 10s, 12s
	timeoutArray := [7]time.Duration{500, 1000, 1500, 4000, 7000, 10000, 12000}
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

		// Attempt transmission
		err := h.sendConfig(buffer[0])

		// If transmission failed, block transmission
		if err != nil {
			log.Printf("Transmission failed: Unable to connect with target with IP %s. Must retry later.\n", h.clientIP)

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

func (h TxHandler) sendConfig(payload []byte) error {
	path := "/a"

	co, err := tcp.Dial(h.clientIP)
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

func (h TxHandler) sendConfigChangeRequest(animals []string) {
	// Save list as a json array before resending it
	buf, err := json.Marshal(animals)
	if err != nil {
		log.Printf("Error config handed to transmittor was faulty: %v\n", err)
	}
	// put buffer into channel
	h.newChan <- buf

}
