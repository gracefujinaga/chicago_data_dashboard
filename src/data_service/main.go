package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"database/sql"

	"github.com/kelvins/geocoder"
	_ "github.com/lib/pq"
)

// Declare my database connection
var db *sql.DB

func init() {
	var err error

	// TODO: need to create the database here

	// Establish connection to Postgres Database

	// OPTION 1 - Postgress application running on localhost
	//db_connection := "user=postgres dbname=chicago_business_intelligence password=root host=localhost sslmode=disable port = 5432"
	// db_connection := "user=postgres dbname=chicago_db password=root host=localhost sslmode=disable"

	// OPTION 2
	// Docker container for the Postgres microservice - uncomment when deploy with host.docker.internal
	//db_connection := "user=postgres dbname=chicago_business_intelligence password=root host=host.docker.internal sslmode=disable port = 5433"

	// OPTION 3
	// Docker container for the Postgress microservice - uncomment when deploy with IP address of the container
	// To find your Postgres container IP, use the command with your network name listed in the docker compose file as follows:
	// docker network inspect cbi_backend
	//db_connection := "user=postgres dbname=chicago_business_intelligence password=root host=162.123.0.9 sslmode=disable port = 5433"

	//Option 4
	//Database application running on Google Cloud Platform.
	db_connection := "user=postgres dbname=chicago_db password=root host=/cloudsql/assignment6-project-441620:us-central1:mypostgres sslmode=disable"

	db, err = sql.Open("postgres", db_connection)
	if err != nil {
		log.Fatal(fmt.Println("Couldn't Open Connection to database"))
		panic(err)
	}

	//Test the database connection
	err = db.Ping()
	if err != nil {
		fmt.Println("Couldn't Connect to database")
		panic(err)
	}

	create_table := `CREATE TABLE IF NOT EXISTS "taxi_trips" (
		"id"   SERIAL , 
		"trip_id" VARCHAR(255) UNIQUE, 
		"trip_start_timestamp" TIMESTAMP WITH TIME ZONE, 
		"trip_end_timestamp" TIMESTAMP WITH TIME ZONE, 
		"pickup_centroid_latitude" DOUBLE PRECISION, 
		"pickup_centroid_longitude" DOUBLE PRECISION, 
		"dropoff_centroid_latitude" DOUBLE PRECISION, 
		"dropoff_centroid_longitude" DOUBLE PRECISION, 
		"pickup_zip_code" VARCHAR(255), 
		"dropoff_zip_code" VARCHAR(255), 
		PRIMARY KEY ("id") 
	);`

	_, _err := db.Exec(create_table)
	if _err != nil {
		panic(_err)
	}

	queryString := `
		SELECT COUNT(*) as length
		FROM taxi_trips;`

	// Variable to hold the count
	var length int

	// Execute the query and scan the result into the length variable
	err = db.QueryRow(queryString).Scan(&length)
	if err != nil {
		log.Fatalf("Failed to query database: %v", err)
	}

	log.Print("getting first batch of data ...")

	// todo: commment back in but im at 95% of my google cloud budget
	geocoder.ApiKey = "AIzaSyCDhgH3J7Utkk_WbKJyKI_Wox4SziNh7JU"

	// get the right data
	fetch_ccvi(db)
	fetch_demographics(db)
	fetch_permits(db)
	fetch_covid(db)

	log.Print("length: ", length)

	if length < 1200 {
		log.Print("fetching transportation data")
		fetch_transportation_paginated()
	}
}

func main() {
	log.Print("starting CBI Microservices ...")

	// Determine port for HTTP service.
	log.Print("starting server...")

	//TODO: add other function handlers here
	http.HandleFunc("/", handler)
	http.HandleFunc("/req4", req_4_handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatal(err)
	}

	// get a new set of permits every day
	for {
		fetch_permits(db)
		time.Sleep(24 * time.Hour)
	}
}
