# Temperature By ZipCode System

## Running the project

1. Clone the repository.

2. Build and run the project using Docker Compose:
```sh
docker-compose up --build

Services
Service A: Receives a CEP and forwards the request to Service B.
Service B: Receives a CEP, finds the location, retrieves the temperature, and returns the formatted response.

Endpoints
Service A:
POST /zipcode: Accepts a JSON body with a "cep" field (string of 8 digits). Returns the city and temperature information.

Service B:
GET /temperature/{zipcode}: Accepts a Route param "zipcode" field (string of 8 digits), finds the city and temperature, and returns the data.

Tracing
This project uses OpenTelemetry and Zipkin for distributed tracing. Zipkin UI is available at http://localhost:9411
