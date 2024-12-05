# Use a specific Python version as the base image
FROM python:3.10

# Set environment variables to prevent Python from writing .pyc files and to enable buffering
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# Set the working directory in the container
WORKDIR /app

# Copy the requirements file into the container
COPY requirements.txt .

# Install system dependencies required for building Python packages
RUN apt-get update && apt-get install -y \
    build-essential \
    libssl-dev \
    libffi-dev \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
RUN pip install --no-cache-dir --upgrade pip 

RUN pip install --upgrade setuptools
RUN pip install cython==3.0.11
RUN pip install numpy==1.26.4
RUN pip install matplotlib==3.9.3
RUN pip install pystan==2.19.1.1
RUN pip install prophet==1.1.6

RUN pip install -r requirements.txt

# Copy the rest of the application code
COPY . .

EXPOSE 5000

ENV go_microservice_url="https://go-microservice-550412521327.us-central1.run.app/"

# Specify the default command to run your application
CMD [ "python", "main.py"]
