from flask import Flask, render_template
import logging
import pandas as pd
import plotly.express as px
import plotly.graph_objects as go
from prophet import Prophet
from prophet.plot import plot_plotly, plot_components_plotly
from flask import Flask, jsonify
from flask import Flask, request, jsonify
import requests
import pandas as pd
import plotly.subplots as sp
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

# route for zip codes
@app.route('/trips_by_zip', methods=['Get'])
def trips_by_zip():
    zipcode = request.args.get('zipcode', None)

    if not zipcode:
        return render_template('no_zip.html', zipcode=zipcode)

     # Fetch data
    response = requests.get(f'{base_url}req4')
    if response.status_code != 200:
        print('Error fetching data')
        exit()
 
    # Create a DataFrame from API response
    data = response.json()
    df = pd.DataFrame(data)

    if zipcode not in df['pickup_zip_code'].values and zipcode not in df['dropoff_zip_code'].values:
        return render_template('no_zip.html', zipcode=zipcode)

    df['trip_start_timestamp'] = pd.to_datetime(df['trip_start_timestamp'], utc=True)
    df['trip_end_timestamp'] = pd.to_datetime(df['trip_end_timestamp'], utc=True)
    df['date'] = df['trip_start_timestamp'].dt.date

    dropoff_df = df[df['dropoff_zip_code'] == zipcode].copy()
    pickup_df = df[df['pickup_zip_code'] == zipcode].copy()

    dropoff_df['type'] = 'Dropoff'
    pickup_df['type'] = 'Pickup'

    # get plotting data
    dropoff_count_df = dropoff_df.groupby(['date'])['trip_id'].count().reset_index(name='Total_Trips')
    pickup_count_df = pickup_df.groupby(['date'])['trip_id'].count().reset_index(name='Total_Trips')
    combined_df = pd.concat([dropoff_df, pickup_df], ignore_index=True)
    combined_count_df = combined_df.groupby(['date', 'type'])['trip_id'].count().reset_index(name='Total_Trips')

    # Create a subplot with 3 rows
    fig = sp.make_subplots(rows=3, cols=1, shared_xaxes=True, vertical_spacing=0.1,
                        subplot_titles=[f'Dropoffs in Zipcode {zipcode}', 
                                        f'Pickups in Zipcode {zipcode}', 
                                        f'Dropoffs and Pickups in Zipcode {zipcode}'])

    # add plots to subplot
    fig.add_trace(go.Scatter(x=dropoff_count_df['date'], y=dropoff_count_df['Total_Trips'], mode='lines+markers', name='Dropoffs', line=dict(color='blue')), row=1, col=1)
    fig.add_trace(go.Scatter(x=pickup_count_df['date'], y=pickup_count_df['Total_Trips'], mode='lines+markers', name='Pickups', line=dict(color='green')), row=2, col=1)
    fig.add_trace(go.Scatter(x=combined_count_df['date'], y=combined_count_df['Total_Trips'], mode='lines+markers', name='Combined', line=dict(color='red')), row=3, col=1)

    # Update layout
    fig.update_layout(title_text=f'Trip Counts for Zipcode {zipcode}', 
                    showlegend=True, 
                    height=900)

    # Save the figure as an HTML file
    figure_html = fig.to_html(full_html=False)

    return render_template('analysis.html', 
                           page_title=f'Trips By Zipcode: {zipcode}',
                           graph=figure_html
                           )

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

    df['time'] = df['trip_start_timestamp'].dt.tz_localize(None)

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
    trip_count_df = df.groupby(['time'])['id'].count().reset_index(name='Total_Trips')
    plot_df = trip_count_df.rename(columns={'time': 'ds', 'Total_Trips': 'y'})

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
    return render_template('forecast_page.html',
                           page_title = title,
                           forecast_title = f'Forecast for {title} Zipcodes',
                           barplot_title = f'Bar Plot for {title} Zipcodes',
                           components_title=f'Components for {title} Zipcodes',
                           graph_html=graph_html,
                           forecast_html=forecast_html,
                           components_html=components_html)

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)