package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"encoding/json"

	_ "github.com/lib/pq"

	_ "github.com/lib/pq"
)

/*
Helper
*/
// PaginationConfig holds configuration for paginated API calls
type PaginationConfig struct {
	BaseURL string                  // Base API URL
	Limit   int                     // Rows per request
	Process func(data []byte) error // Callback to process each batch of data
}

// PaginateAPI performs paginated API requests concurrently
func PaginateAPI(config PaginationConfig, workers int) error {
	fmt.Println("hit paginate api ")

	offset := 0
	var wg sync.WaitGroup
	dataCh := make(chan []byte, workers)
	errCh := make(chan error, workers)
	doneCh := make(chan struct{})

	// Worker pool to process data
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for data := range dataCh {
				if err := config.Process(data); err != nil {
					errCh <- err
					return
				}
			}
		}()
	}

	fmt.Println("right before go func to wait")

	// Close the channels when all workers are done
	go func() {
		wg.Wait()
		close(errCh)
		close(doneCh)
	}()

	fmt.Println("before fetching ALL data")

	// Fetch paginated data
	// limit to 10,000 rows

	for {

		// TODO: can change this as necessary
		// change this back to 1000
		if offset >= 11000 {
			break
		}

		// TODO: fix this back to this uncommented line after
		url := fmt.Sprintf("%s?$limit=%d&$offset=%d", config.BaseURL, config.Limit, offset)
		//url := fmt.Sprintf("%s?$limit=%d&$offset=%d", config.BaseURL, 1000, offset)
		fmt.Printf("Fetching: %s\n", url)

		tr := &http.Transport{
			MaxIdleConns:          10,
			IdleConnTimeout:       1000 * time.Second,
			TLSHandshakeTimeout:   1000 * time.Second,
			ExpectContinueTimeout: 1000 * time.Second,
			DisableCompression:    true,
			Dial: (&net.Dialer{
				Timeout:   1000 * time.Second,
				KeepAlive: 1000 * time.Second,
			}).Dial,
			ResponseHeaderTimeout: 1000 * time.Second,
		}

		client := &http.Client{Transport: tr}

		// Make the HTTP request
		resp, err := client.Get(url)

		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		// Break if no more data
		if len(body) == 0 {
			break
		}

		// Send the data to the processing channel
		dataCh <- body
		offset += config.Limit
	}

	fmt.Println("after fetching ALL data")

	// Close the data channel and wait for workers to finish
	close(dataCh)
	<-doneCh

	// Check if any errors occurred during processing
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

/*
*********************************************
Taxi Trips
**********************************************
*/

func fetch_transportation_paginated() {
	fmt.Println("starting fetch")
	drop_table := `drop table if exists trips_all`
	_, err := db.Exec(drop_table)
	if err != nil {
		panic(err)
	}

	create_table := `CREATE TABLE IF NOT EXISTS "trips_test_client" (
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

	taxiConfig := PaginationConfig{
		//BaseURL: "https://data.cityofchicago.org/resource/wrvz-psew.json",
		BaseURL: "https://data.cityofchicago.org/resource/ajtu-isnz.json",
		Limit:   1000,
		Process: processTaxiTrips,
	}

	rideshareConfig := PaginationConfig{
		//BaseURL: "https://data.cityofchicago.org/resource/m6dm-c72p.json",
		BaseURL: "https://data.cityofchicago.org/resource/n26f-ihde.json",
		Limit:   1000,
		Process: processTaxiTrips,
	}

	// Use 5 concurrent workers for pagination
	if err := PaginateAPI(taxiConfig, 5); err != nil {
		fmt.Printf("Error during pagination: %v\n", err)
	} else {
		fmt.Println("Finished processing trips data.")
	}

	if err := PaginateAPI(rideshareConfig, 5); err != nil {
		fmt.Printf("Error during pagination: %v\n", err)
	} else {
		fmt.Println("Finished processing trips data.")
	}
}

func processTaxiTrips(data []byte) error {

	fmt.Println("starting process function")

	// Unmarshal JSON data into a struct
	var taxi_trips_list TaxiTripsJsonRecords
	if err := json.Unmarshal(data, &taxi_trips_list); err != nil {
		return fmt.Errorf("failed to unmarshal transportation data: %v", err)
	}

	for i := 0; i < len(taxi_trips_list); i++ {

		trip_id := taxi_trips_list[i].Trip_id
		if trip_id == "" {
			continue
		}

		// if trip start/end timestamp doesn't have the length of 23 chars in the format "0000-00-00T00:00:00.000"
		// skip this record
		// get Trip_start_timestamp
		trip_start_timestamp := taxi_trips_list[i].Trip_start_timestamp
		if len(trip_start_timestamp) < 23 {
			continue
		}

		// get Trip_end_timestamp
		trip_end_timestamp := taxi_trips_list[i].Trip_end_timestamp
		if len(trip_end_timestamp) < 23 {
			continue
		}

		pickup_centroid_latitude := taxi_trips_list[i].Pickup_centroid_latitude

		if pickup_centroid_latitude == "" {
			continue
		}
		pickup_centroid_longitude := taxi_trips_list[i].Pickup_centroid_longitude

		if pickup_centroid_longitude == "" {
			continue
		}

		dropoff_centroid_latitude := taxi_trips_list[i].Dropoff_centroid_latitude
		if dropoff_centroid_latitude == "" {
			continue
		}

		dropoff_centroid_longitude := taxi_trips_list[i].Dropoff_centroid_longitude

		if dropoff_centroid_longitude == "" {
			continue
		}

		// dropoff_zip_code, err := GetZipCode(dropoff_centroid_latitude, dropoff_centroid_longitude)
		// if err != nil {
		// 	continue
		// }

		// pickup_zip_code, err := GetZipCode(pickup_centroid_latitude, pickup_centroid_longitude)
		// if err != nil {
		// 	continue
		// }

		pickup_zip_code := "11234"
		dropoff_zip_code := "11234"

		sql := `INSERT INTO trips_test_client ("trip_id", "trip_start_timestamp", "trip_end_timestamp", "pickup_centroid_latitude", "pickup_centroid_longitude", "dropoff_centroid_latitude", "dropoff_centroid_longitude", "pickup_zip_code", 
			"dropoff_zip_code") values($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (trip_id) DO NOTHING;`

		_, err := db.Exec(
			sql,
			trip_id,
			trip_start_timestamp,
			trip_end_timestamp,
			pickup_centroid_latitude,
			pickup_centroid_longitude,
			dropoff_centroid_latitude,
			dropoff_centroid_longitude,
			pickup_zip_code,
			dropoff_zip_code)

		if err != nil {
			panic(err)
		}
	}

	fmt.Println("after inserting data to the API")
	return nil
}
