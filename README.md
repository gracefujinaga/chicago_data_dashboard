# chicago_data_dashboard
MSDS 432 Final Project


# Requirement 1: 
Program and services implemented:
* Data collection and database population for static data
* Data collection and database population for continuously updated data
* Scheduler
* Data Processing and ETL
* PostgreSQL Database
* Analytics Service (shell)
* API endpoint service
* Container Registry

# Requirement 2:
Steps needed to install:
To run locally:
1. set up the database: psql -U postgres -f /Users/gracefujinaga/Documents/Northwestern/MSDS_432/chicago_data_app/src/data_lake_creation/create_tables.sql
2. cd src/data_service
3. go run .
4. navigate to localhost:8080 and type in the url ie localhost:8080/req1 to get the data output
5. cd /Users/gracefujinaga/Documents/Northwestern/MSDS_432/chicago_data_app/src/analysis_service
6. python analysis.py
7. navigate to the specified url and put in the specified url

Steps need to deploy to GCP:
Follow the steps in the readme.pdf here. It is directly from MSDS 432 and Professor Bader created it. Some of the project id and number will be different. 

To access the resource go to:
https://go-microservice-550412521327.us-central1.run.app/

You can type in req1, req2, req3,... req9?zip=60607 after the base url to see the project in action.

If you want to add in the analysis layer, adjust the base url in /Users/gracefujinaga/Documents/Northwestern/MSDS_432/chicago_data_app/src/analysis_service/analysis.py and run 'python analysis.py' from the command line. 
Work from the locally hosted site from there. 


# Running the docker container for the flask app
docker build -t flask-app .

docker run -p 5000:5000 flask-app


# For Grace:

## Todo items:
* debug deploying to the cloud with transportation

Data pull:
* add in pulling the 2024 data: plan would be to keep the old database, then append the new data to the old one
* right now, the json limit is 1000. There are way more rows than that to deal with - probably need to do some type of paging
* use go concurrency to speed up the handling and processing of data

Stretch:
* get the newer data and update the database
    * taxi trips 
    * the other one?
* community area, community area names?


cloud run:
* https://cloud.google.com/run/docs/quickstarts/build-and-deploy/deploy-python-service