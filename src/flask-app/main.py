from flask import Flask, render_template
from datetime import datetime

# from helper_v4 import forecastr,determine_timeframe,get_summary_stats,validate_model,preprocessing
import logging
import time

import pandas as pd

import datetime
from datetime import datetime, date, timedelta
import time
from math import isnan
import numpy as np

import seaborn as sns

import matplotlib.pyplot as plt
import matplotlib.dates as mdates

import plotly.express as px
import plotly.graph_objects as go
from plotly.subplots import make_subplots

from prophet import Prophet
from prophet.diagnostics import cross_validation
from prophet.diagnostics import performance_metrics
from prophet.plot import plot_cross_validation_metric
from prophet.plot import plot_plotly, plot_components_plotly


from flask import Flask, jsonify
from flask import Flask, request, jsonify
import requests
import pandas as pd

import os

import logging

# setup logging
logging.basicConfig(level=logging.INFO)

# get url from the docker container
go_microservice_url = os.getenv("go_microservice_url")
base_url = go_microservice_url
print(base_url)


app = Flask(__name__)

# Route for the homepage
@app.route('/', methods=['GET'])
def home():
    app.logger.info('Home page accessed')
    return render_template('homepage.html')

@app.route('/test', methods=['GET'])
def test():
    app.logger.info('Test page accessed')

    response = requests.get(f'{base_url}req4')
    if response.status_code != 200:
        return jsonify({'error': 'Failed to fetch data from Go service'}), 500

    data = response.json()
    df = pd.DataFrame(data) 
    return data

@app.route('/dropoffs', methods=['GET'])
def dropoffs():

    zipcode = request.args.get('zipcode', None)
    return create_forecast_page('dropoff_zip_code', 'Dropoff', zipcode)

@app.route('/pickups', methods=['GET'])
def pickups():
    zipcode = request.args.get('zipcode', None)
    return create_forecast_page('pickup_zip_code', 'Pickup', zipcode)

@app.route('/all', methods=['GET'])
def all_combined():
    app.logger.info('all page accessed')
    zipcode = request.args.get('zipcode', None)
    return create_forecast_page('zipcode', 'Both Dropoff and Pickup', zipcode)

def create_forecast_page(grouping_col, title, zipcode = None):
    response = requests.get(f'{base_url}req4')
    if response.status_code != 200:
        return jsonify({'error': 'Failed to fetch data from Go service'}), 500

    data = response.json()
    df = pd.DataFrame(data)

    df['trip_start_timestamp'] = pd.to_datetime(df['trip_start_timestamp'], utc=True)
    df['trip_end_timestamp'] = pd.to_datetime(df['trip_end_timestamp'], utc=True)

    df.set_index('trip_start_timestamp', inplace=True)

    df['year'] = df.index.year
    df['month'] = df.index.month
    df['day'] = df.index.day
    df['week_of_year'] = df.index.isocalendar().week
    df['date'] = df.index.date

    if grouping_col == 'zipcode':
        # Duplicate rows for dropoff and pickup
        dropoff_df = df.copy()
        dropoff_df['zipcode'] = dropoff_df['dropoff_zip_code']

        pickup_df = df.copy()
        pickup_df['zipcode'] = pickup_df['pickup_zip_code']

        # Combine into a single DataFrame
        df = pd.concat([dropoff_df, pickup_df], ignore_index=True)

    if zipcode:
        if zipcode not in df['pickup_zip_code'].values and zipcode not in df['dropoff_zip_code'].values:
            err_string = f"zipcode ({zipcode}) does not exist"
            return jsonify({'error': err_string}), 500
        df = df.loc[df[grouping_col] == zipcode]

    counts_df = df.groupby([grouping_col])['trip_id'].count().reset_index(name='trip_count')

     # Create a Plotly bar plot
    fig = px.bar(counts_df, x=grouping_col, y='trip_count', 
                 labels={grouping_col: 'Zip Code', 'trip_count': 'Trip Count'},
                 title=f'Count of Trips by {title} Zip Code')

    # Return the plot in HTML format to embed in the template
    graph_html = fig.to_html(full_html=False)

    # Group by date and calculate total trips
    trip_count_df = df.groupby(['date'])['trip_id'].count().reset_index(name='Total_Trips')
    plot_df = trip_count_df.rename(columns={'date': 'ds', 'Total_Trips': 'y'})

    # Train the Prophet model
    model = Prophet(yearly_seasonality=True, daily_seasonality=True)
    model.fit(plot_df)

    # Create future dates and make predictions
    future_dates = model.make_future_dataframe(periods=50, freq='W')
    forecast = model.predict(future_dates)

    # Prepare Plotly figure for forecast
    forecast_fig = plot_plotly(model, forecast)
    forecast_html = forecast_fig.to_html(full_html=False)

    components_fig = plot_components_plotly(model, forecast)
    components_html = components_fig.to_html(full_html=False)

    # Return the chart to the template
    return render_template('forecast.html',
                           page_title = title,
                           forecast_title = f'Forecast for {title} Zipcodes',
                           barplot_title = f'Bar Plot for {title} Zipcodes',
                           components_title=f'Components for {title} Zipcodes',
                           graph_html=graph_html,
                           forecast_html=forecast_html,
                           components_html=components_html)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5004)
    #app.run(host='0.0.0.0', port=5004)