package main

import (
	"bufio"
	"log"
	"os"
	"reflect"
	"time"
)

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
