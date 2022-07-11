package main

import (
	"log"
)

type RxHandler struct {
	newChan        chan Animal
	currentFogNode string
}

func startAndRunNewRxHandler(dataSourceName string) RxHandler {
	nh := RxHandler{
		newChan:        make(chan Animal, 5),
		currentFogNode: "",
	}

	go nh.startHandler(dataSourceName)

	return nh
}

func (h *RxHandler) startHandler(dataSourceName string) {
	// Connect to database; if database doesn't exist, then exit with error
	animalDB, err := NewAnimalDBInstance(dataSourceName)
	if err != nil {
		log.Fatalln(err)
	}
	for {
		// Wait for the next message.
		animal := <-h.newChan

		// Parse element from json object
		//var animal Animal
		//err := json.Unmarshal(element, &animal)
		if err != nil {
			log.Fatalln(err)
		}

		// add to element to database
		err = animalDB.InsertRow(animal)
		if err != nil {
			log.Println("WARNING: Following error occurred while trying to insert animal to database")
			log.Println(err)

		}
	}
}

func (h RxHandler) handleRequest(animal Animal) {
	// put buffer into channel
	h.newChan <- animal
}
