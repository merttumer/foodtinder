
# FoodTinder

FoodTinder is a food rating app inspired by Tinder. Instead of swiping on people, users vote on their favorite foods on a scale of 1 to 5.

## Table of Contents
- [Features](#features)
- [Environment Variables](#environment-variables)
- [Setup and Installation](#setup-and-installation)
- [Running the App](#running-the-app)
- [Testing the app](#testing-the-app)

## Features
- Users can view various food items.
- Users can rate food items on a scale of 1-5.

## Environment Variables

To run this project, you will need to add the following environment variables when building the Docker image or running the container:

| Variable Name               | Description                               | Default Value |
|-----------------------------|-------------------------------------------|---------------|
| `SERVICE_PORT`              | The port on which the service runs.       | `8080`        |
| `SERVICE_SHUTDOWN_TIMEOUT` | Timeout for service shutdown.              | `15s`         |
| `MONGO_URI`                 | MongoDB connection URI.                   | **Required**  |
| `MONGO_PING_TIMEOUT`        | Timeout for MongoDB ping.                 | `5s`          |
| `MONGO_DATABASE`            | Name of the MongoDB database.             | `foodtinder`  |

## Setup and Installation

1. Clone the repository:

```
git clone https://github.com/merttumer/foodtinder
```

2. Change into the project directory:

```
cd foodtinder
```

3. Build the Docker image:

```
docker build -t foodtinder .
```

4. Start the MongoDB instance if you have it locally or ensure your cloud instance is running.

5. Run the application:

```
docker run -e MONGO_URI=<your-mongo-uri> -p 8080:8080 foodtinder
```

## Running the App

Once you have set up the project, you can run it using the Docker run command mentioned above. Make sure to provide the necessary environment variables as required.

## Testing the App
Once you successfully run the project, you can go into your browser and browse to http://localhost:8080/swagger/index.html