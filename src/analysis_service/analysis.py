from flask import Flask, jsonify
from flask import Flask, request, jsonify
import requests
import pandas as pd

app = Flask(__name__)

# Route for the homepage
@app.route('/', methods=['GET'])
def home():
    return "Welcome to the Chicago Business Intelligence Report!"

# # TODO: update this URL as needed 
base_url = 'http://localhost:8080/'

# base_url = 'https://go-microservice-550412521327.us-central1.run.app/'

@app.route('/req1', methods=['GET'])
def req1():
    response = requests.get(f'{base_url}req1')
    if response.status_code != 200:
        return jsonify({'error': 'Failed to fetch data from Go service'}), 500
    
    data = response.json()

    df = pd.DataFrame(data)
    return data

@app.route('/req2', methods=['GET'])
def req2():
    response = requests.get(f'{base_url}req2')
    if response.status_code != 200:
        return jsonify({'error': 'Failed to fetch data from Go service'}), 500
    
    data = response.json()
    df = pd.DataFrame(data)
    return data


@app.route('/req3', methods=['GET'])
def req3():
    response = requests.get(f'{base_url}req3')
    if response.status_code != 200:
        return jsonify({'error': 'Failed to fetch data from Go service'}), 500
    
    data = response.json()
    df = pd.DataFrame(data)
    return data


@app.route('/req4', methods=['GET'])
def req4():
    response = requests.get(f'{base_url}req4')
    if response.status_code != 200:
        return jsonify({'error': 'Failed to fetch data from Go service'}), 500
    
    data = response.json()
    df = pd.DataFrame(data)
    return data

@app.route('/req5', methods=['GET'])
def req5():
    response = requests.get(f'{base_url}req5')
    if response.status_code != 200:
        return jsonify({'error': 'Failed to fetch data from Go service'}), 500
    
    data = response.json()
    df = pd.DataFrae(data)
    return data


@app.route('/req6', methods=['GET'])
def req6():
    response = requests.get(f'{base_url}req6')
    if response.status_code != 200:
        return jsonify({'error': 'Failed to fetch data from Go service'}), 500
    
    data = response.json()
    df = pd.DataFrame(data)

    return data


@app.route('/req9', methods=['GET'])
def req9():
    response = requests.get(f'{base_url}req9')
    if response.status_code != 200:
        return jsonify({'error': 'Failed to fetch data from Go service'}), 500
    
    data = response.json()
    df = pd.DataFrame(data)

    return data


if __name__ == '__main__':
    app.run(port=5000, debug=True)
