package main

import (
	"encoding/json" // PostgreSQL driver
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

// Function to query the database with the whole query string passed in
func queryDatabase(queryString string, args ...interface{}) ([]map[string]interface{}, error) {

	// db_connection := "user=postgres dbname=chicago_db password=root host=/cloudsql/assignment6-project-441620:us-central1:mypostgres sslmode=disable"

	// db, err := sql.Open("postgres", db_connection)
	// if err != nil {
	// 	log.Fatal(fmt.Println("Couldn't Open Connection to database - after opening database"))
	// 	panic(err)
	// }

	//Test the database connection
	err := db.Ping()
	if err != nil {
		fmt.Println("Couldn't Connect to database")
		panic(err)
	}

	// Execute the query with the provided parameters
	rows, err := db.Query(queryString, args...)
	if err != nil {
		return nil, fmt.Errorf("could not execute query: %v", err)
	}
	defer rows.Close()

	// Parse the rows into a slice of maps
	var results []map[string]interface{}
	for rows.Next() {
		// Use a slice of interfaces to hold row data, for flexible handling of columns
		columns, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("could not get columns: %v", err)
		}
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("could not scan row: %v", err)
		}

		// Create a map to store the column names and their respective values
		result := make(map[string]interface{})
		for i, colName := range columns {
			result[colName] = values[i]
		}
		results = append(results, result)
	}

	// Check for any error during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %v", err)
	}

	return results, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving:", r.URL.Path, "from", r.Host)
	w.WriteHeader(http.StatusOK)
	Body := "Thanks for visiting!\n"
	fmt.Fprintf(w, "%s", Body)
}

func req_1_handler(w http.ResponseWriter, r *http.Request) {

	query := `
		-- Step 1: Categorize COVID-19 cases into low, medium, and high
			WITH categorized_covid AS (
				SELECT
					zipcode,
					DATE_TRUNC('week', week_start AT TIME ZONE 'UTC') AS week_start,
					cases_weekly,
					CASE
						WHEN cases_weekly < 10 THEN 'low'
						WHEN cases_weekly BETWEEN 10 AND 150 THEN 'medium'
						ELSE 'high'
					END AS covid_category
				FROM covid
			),

			-- Step 2: Aggregate taxi trips by week and zip code
			trips_by_week AS (
				SELECT
					DATE_TRUNC('week', trip_start_timestamp) AS trip_week,
					pickup_zip_code AS zip_code,
					COUNT(*) AS num_trips
				FROM trips
				GROUP BY trip_week, pickup_zip_code
				UNION ALL
				SELECT
					DATE_TRUNC('week', trip_start_timestamp) AS trip_week,
					dropoff_zip_code AS zip_code,
					COUNT(*) AS num_trips
				FROM trips
				GROUP BY trip_week, dropoff_zip_code
			)

				SELECT
					t.trip_week,
					t.zip_code,
					SUM(t.num_trips) AS total_trips,
					c.covid_category
				FROM trips_by_week t
				JOIN categorized_covid c
					ON t.zip_code = c.zipcode
					AND t.trip_week = c.week_start
				GROUP BY t.trip_week, t.zip_code, c.covid_category
	`
	// Fetch the data from the database using the provided query string
	data, err := queryDatabase(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response header as JSON and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Format the result into JSON and send the response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
	}
}

// handler for req 2 dataset
func req_2_handler(w http.ResponseWriter, r *http.Request) {

	// Get the raw query string from the request body or URL
	queryString := `WITH categorized_trips AS (
    SELECT
        CASE
            WHEN t."pickup_zip_code" = '60638' THEN 'Midway'
            WHEN t."pickup_zip_code" = '60666' THEN 'OHare'
            ELSE 'Other'
        END AS "pickup_location",
        t."dropoff_zip_code",
        COUNT(t."id") AS "trip_count"
    FROM
        "trips" t
    WHERE
        t."pickup_zip_code" IN ('60638', '60656', '60666')  -- Filter for Midway and O'Hare
    GROUP BY
        "pickup_location", t."dropoff_zip_code"
	)
	SELECT
		"pickup_location",
		"dropoff_zip_code",
		CAST(SUM("trip_count") AS TEXT) AS "total_trips"
	FROM
		categorized_trips
	GROUP BY
		CUBE("pickup_location", "dropoff_zip_code")
	ORDER BY
		"pickup_location", "dropoff_zip_code";`

	// Fetch the data from the database using the provided query string
	data, err := queryDatabase(queryString)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response header as JSON and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Format the result into JSON and send the response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
	}
}

func req_3_handler(w http.ResponseWriter, r *http.Request) {
	query := `
		WITH high_ccvi_neighborhoods AS (
		SELECT
			c."zipcode",
			c."community_area",
			c."community_area_name"
		FROM
			"ccvi" c
		WHERE
			c."category" = 'HIGH'
		)
		SELECT
			t."pickup_zip_code",
			t."dropoff_zip_code",
			COUNT(t."id") AS "trip_count"
		FROM
			"trips" t
		JOIN
			high_ccvi_neighborhoods h ON t."pickup_zip_code" = h."zipcode" 
									OR t."dropoff_zip_code" = h."zipcode"
		JOIN
			"covid" c ON t."pickup_zip_code" = c."zipcode" 
					OR t."dropoff_zip_code" = c."zipcode"
		GROUP BY
			CUBE(t."pickup_zip_code", t."dropoff_zip_code")
		ORDER BY
			t."pickup_zip_code", t."dropoff_zip_code";
	`

	// Fetch the data from the database using the provided query string
	data, err := queryDatabase(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response header as JSON and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Format the result into JSON and send the response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
	}
}

func req_4_handler(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT *
		FROM taxi_trips
	`

	// Fetch the data from the database using the provided query string
	data, err := queryDatabase(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response header as JSON and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Format the result into JSON and send the response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
	}
}

func req_5_handler(w http.ResponseWriter, r *http.Request) {
	query := `
			WITH top_neighborhoods AS (
				SELECT 
					"community_area",
					"community_area_name",
					"unemployment",
					"below_poverty_level"
				FROM 
					"demographics"
				WHERE
					community_area is NOT NULL 
				ORDER BY 
					"unemployment" DESC, 
					"below_poverty_level" DESC
				LIMIT 5
			),
			filtered_permits AS (
				SELECT *
				FROM "building_permits" b 
				WHERE b."community_area" != 'NaN'
			)

			SELECT 
					b."community_area",
					DATE_TRUNC('month', b."issue_date") AS permit_month,
					COUNT(b."id") AS permit_count
			FROM 
				"filtered_permits" b
			INNER JOIN 
				top_neighborhoods t ON b."community_area"::INT = t."community_area"
			GROUP BY 
				b."community_area", DATE_TRUNC('month', b."issue_date")

	`

	// Fetch the data from the database using the provided query string
	data, err := queryDatabase(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response header as JSON and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Format the result into JSON and send the response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
	}
}

func req_6_handler(w http.ResponseWriter, r *http.Request) {
	// TODO
	// note that this query will not return anything because the community areas
	// with permit-new construction all have Nan Zip codes and nan community area numbers

	query := `
			SELECT b."community_area", COUNT(b."id") AS permit_count
			FROM "building_permits" b 
				INNER JOIN "demographics" d on b."community_area"::INT = d."community_area"
			WHERE b."community_area" != 'NaN' 
				AND b.permit_type = 'PERMIT - NEW CONSTRUCTION'
				AND d."per_capita_income" < 30000
		GROUP BY 
				b."community_area"
			ORDER BY 
				b."community_area"
	`

	// Fetch the data from the database using the provided query string
	data, err := queryDatabase(query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response header as JSON and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Format the result into JSON and send the response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
	}
}

func req_9_handler(w http.ResponseWriter, r *http.Request) {

	zipcode := r.URL.Query().Get("zip")

	query := "SELECT * FROM trips"
	var args []interface{}

	if zipcode != "" {
		query += " WHERE pickup_zip_code = $1 OR dropoff_zip_code = $1"
		args = append(args, zipcode)
	}

	// Fetch the data from the database using the provided query string
	data, err := queryDatabase(query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response header as JSON and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Format the result into JSON and send the response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
	}
}
