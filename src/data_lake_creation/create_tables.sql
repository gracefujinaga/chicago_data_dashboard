DROP DATABASE IF EXISTS chicago_db;
CREATE DATABASE chicago_db;

\c chicago_db;

DROP TABLE IF EXISTS transportation;
DROP TABLE IF EXISTS covid;
DROP TABLE IF EXISTS ccvi;
DROP TABLE IF EXISTS building_permits;
DROP TABLE IF EXISTS demographics;


CREATE TABLE transportation (
    trip_id TEXT,
    pickup_community_area INT,
    dropoff_community_area INT,
    trip_start_timestamp TIMESTAMP(3),
    drop_off_zipcode INT,
    pickup_zipcode INT
);


CREATE TABLE covid (
    zipcode INT,
    week_start TIMESTAMP(3),
    test_rate_weekly FLOAT,
    precent_tested_pos_weekly FLOAT,
    cases_weekly INT
    -- do I need community area?
);

CREATE TABLE ccvi (
    geography_type TEXT,
    community_area INT,
    zipcode INT,
    ccvi_category TEXT
);

CREATE TABLE building_permits (
    id TEXT,
    permit_number TEXT,
    permit_type TEXT,
    permit_status TEXT,
    zipcode INT,
    community_area INT
);

CREATE TABLE demographics (
    community_area INT,
    community_area_name TEXT,
    below_poverty_level FLOAT,
    unemployment FLOAT,
    per_capita_income FLOAT
);

-- table for community area number to name?

-- command to set up the database:
-- psql -U postgres -f /Users/gracefujinaga/Documents/Northwestern/MSDS_432/chicago_data_app/src/data_lake_creation/create_tables.sql

-- psql -h localhost -U postgres -d chicago_db