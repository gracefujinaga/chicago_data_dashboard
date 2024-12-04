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
RUN pip install cython
RUN pip install numpy
RUN pip install matplotlib
RUN pip install pystan
RUN pip install prophet

RUN pip install -r requirements.txt

# Copy the rest of the application code
COPY . .

EXPOSE 5000

# Specify the default command to run your application
#CMD ["python", "main.py"]
CMD [ "python", "main.py"]
