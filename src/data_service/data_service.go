package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"encoding/json"

	"github.com/kelvins/geocoder"
	_ "github.com/lib/pq"
)

type TaxiTripsJsonRecords []struct {
	Trip_id                    string `json:"trip_id"`
	Trip_start_timestamp       string `json:"trip_start_timestamp"`
	Trip_end_timestamp         string `json:"trip_end_timestamp"`
	Pickup_centroid_latitude   string `json:"pickup_centroid_latitude"`
	Pickup_centroid_longitude  string `json:"pickup_centroid_longitude"`
	Dropoff_centroid_latitude  string `json:"dropoff_centroid_latitude"`
	Dropoff_centroid_longitude string `json:"dropoff_centroid_longitude"`
}

type DemographicsJsonRecords []struct {
	Community_area      string `json:"community_area"`
	Community_area_name string `json:"community_area_name"`
	Below_poverty_level string `json:"below_poverty_level"`
	Unemployment        string `json:"unemployment"`
	Per_capita_income   string `json:"per_capita_income"`
}

type BuildingPermitJsonRecords []struct {
	PermitNumber           string `json:"permit_"`
	PermitType             string `json:"permit_type"`
	PermitStatus           string `json:"permit_status,omitempty"`
	Application_start_date string `json:"application_start_date"`
	Issue_date             string `json:"issue_date"`
	CommunityArea          string `json:"community_area,omitempty"`
	Latitude               string `json:"latitude,omitempty"`
	Longitude              string `json:"longitude,omitempty"`
}

type CovidJsonStruct []struct {
	Zipcode                 string `json:"zip_code"`
	Week_start              string `json:"week_start"`
	Test_rate_weekly        string `json:"test_rate_weekly"`
	Percent_tested_positive string `json:"percent_tested_positive_weekly"`
	Cases_weekly            string `json:"cases_weekly"`
}

type Location struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type CCVIJsonRecords []struct {
	Geography_type        string   `json:"geography_type"` // "ZIP" or "CA"
	Community_area_or_zip string   `json:"community_area_or_zip"`
	Community_area_name   string   `json:"community_area_name,omitempty"` // empty if community_area_or_zip is a zip
	CCVI_Category         string   `json:"ccvi_category,omitempty"`
	Location              Location `json:"location"`
}

/*
Helper function: takes in two strings
returns zipcode string
*/
func GetZipCode(latitude string, longitude string) (string, error) {
	// Check for empty inputs
	if latitude == "" || longitude == "" {
		return "NaN", fmt.Errorf("latitude or longitude cannot be empty")
	}

	// Convert latitude and longitude to float
	latitudeFloat, err := strconv.ParseFloat(latitude, 64)
	if err != nil {
		return "NaN", fmt.Errorf("invalid latitude: %v", err)
	}

	longitudeFloat, err := strconv.ParseFloat(longitude, 64)
	if err != nil {
		return "NaN", fmt.Errorf("invalid longitude: %v", err)
	}

	// Use the geocoding library to reverse geocode
	location := geocoder.Location{
		Latitude:  latitudeFloat,
		Longitude: longitudeFloat,
	}

	address, err := geocoder.GeocodingReverse(location)
	if err != nil {
		return "NaN", fmt.Errorf("failed to get address from geocoding API: %v", err)
	}

	// Validate address result
	if len(address) == 0 || address[0].PostalCode == "" {
		return "NaN", fmt.Errorf("no postal code found for given coordinates")
	}

	return address[0].PostalCode, nil
}

/*
*********************************************
Demographics
**********************************************
*/
func fetch_demographics(db *sql.DB) {

	drop_table := `drop table if exists demographics`
	_, err := db.Exec(drop_table)
	if err != nil {
		panic(err)
	}

	create_table := `CREATE TABLE IF NOT EXISTS "demographics" (
			"id" SERIAL,
			"community_area" INT,
			"community_area_name" TEXT,
			"below_poverty_level" FLOAT,
			"unemployment" FLOAT,
			"per_capita_income" FLOAT
			);`

	_, _err := db.Exec(create_table)
	if _err != nil {
		panic(_err)
	}

	var url = "https://data.cityofchicago.org/resource/iqnk-2tcu.json?$limit=50"

	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	body, _ := ioutil.ReadAll(res.Body)

	var demographics_list DemographicsJsonRecords
	json.Unmarshal(body, &demographics_list)

	for i := 0; i < len(demographics_list); i++ {

		// get all of the fields
		// only keep the record if it has all of the necessary fields
		community_area := demographics_list[i].Community_area
		if community_area == "" {
			continue
		}

		community_area_name := demographics_list[i].Community_area_name
		if community_area_name == "" {
			continue
		}

		below_poverty_level := demographics_list[i].Below_poverty_level
		if below_poverty_level == "" {
			continue
		}

		unemployment := demographics_list[i].Unemployment
		if unemployment == "" {
			continue
		}

		per_capita_income := demographics_list[i].Per_capita_income
		if per_capita_income == "" {
			continue
		}

		sql := `INSERT INTO demographics ("community_area", "community_area_name", "below_poverty_level", "unemployment", "per_capita_income")
			 values($1, $2, $3, $4, $5)`

		_, err = db.Exec(
			sql,
			community_area,
			community_area_name,
			below_poverty_level,
			unemployment,
			per_capita_income)

		if err != nil {
			panic(err)
		}

	}
}

/*
Building Permits
*/

func fetch_permits(db *sql.DB) {

	fmt.Println("start buildings")

	drop_table := `drop table if exists building_permits`
	_, err := db.Exec(drop_table)
	if err != nil {
		panic(err)
	}

	create_table := `CREATE TABLE IF NOT EXISTS "building_permits" (
		"id" SERIAL,
		"permit_number" TEXT,
		"permit_type" TEXT,
		"permit_status" TEXT,
		"community_area" TEXT,
		"zipcode" TEXT,
		"application_start_date" TIMESTAMP WITH TIME ZONE,
		"issue_date" TIMESTAMP WITH TIME ZONE
	);`
	_, _err := db.Exec(create_table)
	if _err != nil {
		panic(_err)
	}

	var url = "https://data.cityofchicago.org/resource/ydr8-5enu?$limit=50"
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	body, _ := ioutil.ReadAll(res.Body)

	var permitsList BuildingPermitJsonRecords

	// Attempt to unmarshal the JSON
	err = json.Unmarshal(body, &permitsList)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	for i := 0; i < len(permitsList); i++ {
		// Check if permit_number is empty, otherwise assign "NaN"
		if permitsList[i].PermitNumber == "" {
			permitsList[i].PermitNumber = "NaN"
		}

		// Check if permit_type is empty, otherwise assign "NaN"
		if permitsList[i].PermitType == "" {
			permitsList[i].PermitType = "NaN"
		}

		// Check if permit_status is empty, otherwise assign "NaN"
		if permitsList[i].PermitStatus == "" {
			permitsList[i].PermitStatus = "NaN"
		}

		// Check if community_area is empty, otherwise assign "NaN"
		if permitsList[i].CommunityArea == "" {
			permitsList[i].CommunityArea = "NaN"
		}

		// Check if latitude and longitude are empty, otherwise assign empty strings
		if permitsList[i].Latitude == "" {
			permitsList[i].Latitude = ""
		}

		if permitsList[i].Longitude == "" {
			permitsList[i].Longitude = ""
		}

		application_start_date := permitsList[i].Application_start_date
		if len(application_start_date) < 23 || application_start_date == "" {
			fmt.Println("skipped")
			continue
		}

		issue_date := permitsList[i].Issue_date
		if len(issue_date) < 23 || issue_date == "" {
			fmt.Println("skipped")
			continue
		}

		// Use geocoding to find the zipcode
		zipcode, err := GetZipCode(permitsList[i].Latitude, permitsList[i].Longitude)
		if err != nil {
			zipcode = "NaN" // Use "NaN" if geocoding fails
		}

		// Insert into the database
		sql := `INSERT INTO building_permits ("permit_number", "permit_type", "permit_status", "community_area", "zipcode", "application_start_date", "issue_date") 
			values($1, $2, $3, $4, $5, $6, $7)`

		_, err = db.Exec(
			sql,
			permitsList[i].PermitNumber,
			permitsList[i].PermitType,
			permitsList[i].PermitStatus,
			permitsList[i].CommunityArea,
			zipcode,
			permitsList[i].Application_start_date,
			permitsList[i].Issue_date,
		)

		if err != nil {
			fmt.Println("Error inserting into database:", err)
		}
	}

	fmt.Println("end buildings")
}

/*
CCVI DATA
*/
func fetch_ccvi(db *sql.DB) {

	drop_table := `drop table if exists ccvi`
	_, err := db.Exec(drop_table)
	if err != nil {
		panic(err)
	}

	create_table := `CREATE TABLE IF NOT EXISTS "ccvi" (
		"id" SERIAL,
		"zipcode" TEXT,
		"community_area" TEXT,
		"community_area_name" TEXT,
		"category" TEXT
	);`

	_, _err := db.Exec(create_table)
	if _err != nil {
		panic(_err)
	}

	var url = "https://data.cityofchicago.org/resource/xhc6-88s9.json?$limit=50"
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	body, _ := ioutil.ReadAll(res.Body)

	var ccviList CCVIJsonRecords
	err = json.Unmarshal(body, &ccviList)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(ccviList); i++ {

		// get category: if no category, continue
		category := ccviList[i].CCVI_Category
		if category == "" {
			continue
		}

		var zipcode string
		var community_area string
		var community_area_name string

		if ccviList[i].Geography_type == "ZIP" {

			// logic for if community area is ZIP
			zipcode = ccviList[i].Community_area_or_zip
			if zipcode == "" {
				continue
			}

			community_area = "NaN"
			community_area_name = "NaN"

		} else if ccviList[i].Geography_type == "CA" {
			// do logic for community_area
			community_area = ccviList[i].Community_area_or_zip
			if community_area == "" {
				continue
			}

			community_area_name = ccviList[i].Community_area_name
			if community_area_name == "" {
				continue
			}

			// calculate zip code
			latitude := fmt.Sprintf("%f", ccviList[i].Location.Coordinates[1])
			longitude := fmt.Sprintf("%f", ccviList[i].Location.Coordinates[0])

			zipcode, err = GetZipCode(latitude, longitude)
			if err != nil {
				zipcode = "NaN"
			}
		} else {
			continue
		}

		// Insert into the database
		sql := `INSERT INTO ccvi ("zipcode", "community_area", "community_area_name", "category") 
			values($1, $2, $3, $4)`

		_, err = db.Exec(
			sql,
			zipcode,
			community_area,
			community_area_name,
			category,
		)

		if err != nil {
			fmt.Println("Error inserting into database:", err)
			panic(err)
		}

	}
}

func fetch_covid(db *sql.DB) {

	fmt.Println("get covid")

	drop_table := `drop table if exists covid`
	_, err := db.Exec(drop_table)
	if err != nil {
		panic(err)
	}

	create_table := `CREATE TABLE IF NOT EXISTS "covid" (
		"id" SERIAL,
		"zipcode" TEXT,
		"week_start" TIMESTAMP WITH TIME ZONE,
		"test_rate_weekly" DOUBLE PRECISION,
		"percent_tested_positive" DOUBLE PRECISION,
		"cases_weekly" INT
	);`

	_, _err := db.Exec(create_table)
	if _err != nil {
		panic(_err)
	}

	var url = "https://data.cityofchicago.org/resource/yhhz-zm2v.json?$limit=50"
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	body, _ := ioutil.ReadAll(res.Body)

	var covidList CovidJsonStruct
	err = json.Unmarshal(body, &covidList)
	if err != nil {
		panic(err)
		log.Print("failed to unmarshal covidList")
		return
	}

	for i := 0; i < len(covidList); i++ {

		zipcode := covidList[i].Zipcode
		if zipcode == "" {
			continue
		}

		week_start := covidList[i].Week_start
		if len(week_start) < 23 || week_start == "" {
			continue
		}

		test_rate_weekly := covidList[i].Test_rate_weekly
		if test_rate_weekly == "" {
			continue
		}

		percent_tested_positive := covidList[i].Percent_tested_positive
		if percent_tested_positive == "" {
			continue
		}

		cases_weekly := covidList[i].Cases_weekly
		if cases_weekly == "" {
			continue
		}

		// Insert into the database
		sql := `INSERT INTO covid ("zipcode", "week_start", "test_rate_weekly", "percent_tested_positive", "cases_weekly") 
			values($1, $2, $3, $4, $5)`

		_, err = db.Exec(
			sql,
			zipcode,
			week_start,
			test_rate_weekly,
			percent_tested_positive,
			cases_weekly,
		)

		if err != nil {
			fmt.Println("Error inserting into database:", err)
			panic(err)
		}

	}

	fmt.Println("end covid")

}
