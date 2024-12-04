package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
	db_connection := "user=postgres dbname=chicago_db password=root host=localhost sslmode=disable"

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
	//db_connection := "user=postgres dbname=chicago_db password=root host=/cloudsql/assignment6-project-441620:us-central1:mypostgres sslmode=disable"

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

	// todo: commment back in but im at 95% of my google cloud budget
	geocoder.ApiKey = "AIzaSyCDhgH3J7Utkk_WbKJyKI_Wox4SziNh7JU"
}

func main() {
	log.Print("starting CBI Microservices ...")

	// TODO: uncomment
	fetch_ccvi(db)
	fetch_demographics(db)

	// // //functions that use the pagination func
	// fetch_transportation(db)
	// fetch_permits(db)
	// fetch_covid(db)

	fetch_transportation_paginated()

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

	//TODO: uncomment
	// for {
	// 	fetch_permits()
	// 	time.Sleep(24 * time.Hour)
	// }
}
