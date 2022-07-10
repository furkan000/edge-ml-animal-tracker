package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type Animal struct {
	DetectionID    int     `json:"detection_id"`
	CameraUuid     int     `json:"camera_uuid"`
	DetectionTime  string  `json:"detection_time"`
	DetectedAnimal string  `json:"Detected_object"`
	Temperature    float64 `json:"temperature"`
}

type AnimalDB struct {
	database        *sql.DB
	addRowStatement *sql.Stmt
}

// NewAnimalDBInstance creates a new db instance by opening the Database file.
// Assumes that a table exists
func NewAnimalDBInstance(dataSourceName string) (*AnimalDB, error) {

	log.Println("Connect to database instance")
	dbInstance := &AnimalDB{}
	// initialize database assumes initialized dataSourceName
	err := dbInstance.ConnectToDatabase(dataSourceName)
	if err != nil {
		return nil, err
	}

	log.Println("preparing often used query statements for the database")
	// initiate the addRowStatement
	dbInstance.addRowStatement, err =
		dbInstance.database.Prepare(
			"INSERT INTO detected_animals (detection_id, camera_uuid, detection_time, detected_animal, temperature) VALUES (?, ?, ?, ?, ?)")

	if err != nil {
		log.Fatalln(err)
	}

	return dbInstance, nil
}

// ConnectToDatabase opens the database file using the imported sqlite driver
func (db *AnimalDB) ConnectToDatabase(DataSourceName string) error {
	log.Println("Opening database")
	var err error
	db.database, err =
		sql.Open("mysql", DataSourceName)

	if err != nil {
		return err
	}
	return nil
}

// CloseDatabaseConnection closes the database file
func (db *AnimalDB) CloseDatabaseConnection() {
	err := db.database.Close()
	if err != nil {
		log.Fatal("error closing database.")
	}
}

// InsertRow adds animal detections instance into the animals Database
func (db *AnimalDB) InsertRow(animal Animal) error {
	//(time string, customerUsername int, adClicks int, adDownloads int, successRate float32, Comment string) {
	log.Printf(
		"Inserting new detection to database: %s,\t%s,\t%s",
		animal.DetectionTime,
		animal.CameraUuid,
		animal.DetectedAnimal)

	_, err := db.addRowStatement.Exec(
		0,
		animal.CameraUuid,
		animal.DetectionTime,
		animal.DetectedAnimal,
		animal.Temperature)

	return err
}

// GetAllAnimalRows prints animal detections in the animal table into the logs.
func (db *AnimalDB) GetAllAnimalRows() ([]Animal, error) {
	log.Println("Fetching all rows from the animals database.")

	var rows *sql.Rows
	var query *sql.Stmt
	var err error

	query, err = db.database.Prepare(
		"SELECT * FROM detected_animals")
	if err != nil {
		// Query Prep failed
		return nil, err
	}

	rows, err = query.Query()
	if err != nil {
		// Could not execute query. remember to set up the database
		return nil, err
	}

	animals, err := scanAnimalsFromRows(rows)
	if err != nil {
		// Could not execute query. remember to set up the database
		return nil, err
	}

	return animals, nil
}

// scanAnimalsFromRows scans and returns all animal detections in the database.
func scanAnimalsFromRows(rows *sql.Rows) ([]Animal, error) {
	log.Println("Parsing all animal rows to an animal detection array.")

	var animals []Animal

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		animal := Animal{}
		err := rows.Scan(
			&animal.DetectionID,
			&animal.CameraUuid,
			&animal.DetectionTime,
			&animal.DetectedAnimal,
			&animal.Temperature,
		)

		if err != nil {
			return []Animal{}, err
		}

		err = rows.Err()
		if err != nil {
			return []Animal{}, err
		}
		animals = append(animals, animal)
	}

	return animals, nil
}
