package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/plgd-dev/go-coap/v2"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
)

var globalReceiverHandler RxHandler

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

	globalReceiverHandler.handleRequest(animal)

	err = w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("gg fam")))
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

func readAnimalsFromFile(configFile string) ([]string, error) {
	file, err := os.Open(configFile)
	if err != nil {
		log.Printf("Error occured reading config: %v\n", err)
		return nil, err
	}

	var animals []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		animals = append(animals, scanner.Text())
	}

	err = file.Close()
	if err != nil {
		log.Printf("Error occured closing config file: %v\n", err)
		return nil, err
	}

	return animals, nil
}

func watchConfigFile(configFile string, clients []string) {
	if len(clients) == 0 {
		log.Fatalln("Server instance required a list of clients to function.")
	}

	clientsTx := make([]TxHandler, len(clients))
	for i, val := range clients {
		clientsTx[i] = startAndRunNewTxHandler(val)
	}

	go func() {
		curAnimals, err := readAnimalsFromFile(configFile)
		if err != nil {
			log.Fatalln("couldn't read line.")
		}

		if len(curAnimals) > 0 {
			// send animals to all clients
			log.Println("Sending config file to all clients")
			for _, v := range clientsTx {
				v.sendConfigChangeRequest(curAnimals)
			}
		}

		for {
			time.Sleep(10 * time.Second)

			// read config file and send its content to the clients, if it's content is valid
			animals, err := readAnimalsFromFile(configFile)
			if err != nil {
				log.Fatalln("couldn't read line.")
			}

			if !reflect.DeepEqual(curAnimals, animals) {
				log.Println("Config file altered. Read it and update list in all clients.")
				if len(animals) > 0 {
					// send animals to all clients
					log.Println("Sending config file to all clients")
					for _, v := range clientsTx {
						v.sendConfigChangeRequest(animals)
					}
				}
				curAnimals = animals
			}
		}
	}()
}

func main() {
	fogNodePort := os.Getenv("SERVER_HOG_FOG_NODE_PORT")                // Default: ":3444"
	dataSourceName := os.Getenv("SERVER_HOG_DATA_SOURCE_NAME")          // Default: "root:my_fog_password@(172.104.142.115:3306)/my_database"
	configFile := os.Getenv("SERVER_HOG_CONFIG_FILE")                   // Default: "config.txt"
	clientsConfigurationIPString := os.Getenv("SERVER_HOG_CLIENT_LIST") // Default: Default: "[\"localhost:3555\"]"

	if fogNodePort == "" || dataSourceName == "" || configFile == "" || clientsConfigurationIPString == "" {
		log.Printf("Environmental variables not initialized, using default values")

		fogNodePort = ":3444"
		dataSourceName = "root:my_fog_password@(172.104.142.115:3306)/my_database"
		configFile = "config.txt"
		clientsConfigurationIPString = "[\"localhost:3555\"]"
	}

	var clientsConfigurationIP []string
	err := json.Unmarshal([]byte(clientsConfigurationIPString), &clientsConfigurationIP)
	if err != nil {
		log.Fatalf("Issues reading client config: %v\n", err)
	}

	if len(clientsConfigurationIP) == 0 {
		log.Fatalln("Cannot start a server without setting its clients (cameras) statically. We realize that this" +
			"isn't best practice, but that should suffice for a prototype.")
	}

	// watch config file, and send information in case it is altered.
	// The config file contains a list of all the animals, which we're keeping track of.
	go watchConfigFile(configFile, clientsConfigurationIP)

	// start receiver handler (handler that adds items to database)
	globalReceiverHandler = startAndRunNewRxHandler(dataSourceName)

	// start coap listener
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	err = r.Handle("/a", mux.HandlerFunc(handleA))
	if err != nil {
		return
	}

	log.Fatal(coap.ListenAndServe("tcp", fogNodePort, r))
}
