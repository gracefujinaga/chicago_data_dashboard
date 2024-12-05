# chicago_data_dashboard
MSDS 432 Final Project

# Navigating the Repository
- The src directory contains all relevant source code for the microservices
- cloudbuild_versions has all of the cloudbuild files used. They are built as options for deploying just the flask-app or just the go-microservices
- the data_lake_creation has the sql file to set up the local database but isnt utilized in the final version

# Requirement 1: 
**Program and services implemented:**
* Data collection and database population for static data
* Data collection and database population for continuously updated data
* Scheduler
* Data Processing and ETL
* PostgreSQL Database
* API endpoint service
* Frontend web app dashboard

# Requirement 2:
**Steps needed to install and run go-microservices:**
To run locally:
1. set up the database: psql -U postgres -f /Users/gracefujinaga/Documents/Northwestern/MSDS_432/chicago_data_app/src/data_lake_creation/create_tables.sql
2. cd src/data_service
3. go run .
4. navigate to localhost:8080 and type in the url ie localhost:8080/req4 to get the data for requirement 4

**Steps need to deploy to GCP:**
Follow the steps in the readme.pdf here. It is directly from MSDS 432 and Professor Bader created it. Some of the project id and number will be different. 

To access the resource go to:
https://go-microservice-550412521327.us-central1.run.app/

**Steps to run the front end app:**
1. Navigate to the flask-app folder in the terminal
2. docker build -t flask-app .
3. docker run -p 5000:5000 flask-app
4. access the link in the terminal 


