# Go MQTT TimescaleDB Connector

This project connects to an MQTT broker, listens for IoT sensor data, and stores it in a TimescaleDB (PostgreSQL with time-series extension) database.

## Features

- MQTT client that subscribes to a configured topic
- Automatic parsing of sensor data in JSON format
- TimescaleDB storage for efficient time-series data handling
- Automatic table creation if it doesn't exist

## Requirements

- Go 1.19 or higher
- PostgreSQL with TimescaleDB extension
- MQTT broker (e.g., Mosquitto)

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/go-mqtt-timescale.git
   cd go-mqtt-timescale
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

## Configuration

Edit the `config.yaml` file to configure:

- MQTT broker connection details
- PostgreSQL/TimescaleDB connection details
- Table name for sensor data

```yaml
mqtt:
  broker: "https://mqtt.ponytojas.dev"
  port: 8883  # Default port for MQTT over TLS
  client_id: "go-mqtt-client"
  topic: "sensors/data"
  username: "your_username"
  password: "your_password"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "iot_data"
  sslmode: "disable"

timescale:
  table_name: "sensor_data"
```

The broker URL supports several formats:
- `https://mqtt.ponytojas.dev` - HTTPS URL (automatically converted to ssl:// with port 8883)
- `ssl://mqtt.ponytojas.dev:8883` - Direct SSL protocol
- `tcp://mqtt.ponytojas.dev:1883` - Unencrypted TCP protocol

You can also set the broker URL via the environment variable `MQTT_BROKER_URL`.

## Running the Application

```
go run cmd/main.go
```

## Expected JSON Format

The application expects sensor data in the following JSON format:

```json
{
  "timestamp": "2023-05-20T15:04:05Z",
  "temperature": 24.5,
  "humidity": 65.2,
  "light": 850
}
```

- `timestamp`: RFC3339 formatted timestamp (if not provided, current time will be used)
- `temperature`: Temperature reading (float)
- `humidity`: Humidity reading (float)
- `light`: Light intensity reading (float)

## Database Schema

The application creates a TimescaleDB hypertable with the following schema:

```sql
CREATE TABLE sensor_data (
    time TIMESTAMPTZ NOT NULL,
    temperature DOUBLE PRECISION,
    humidity DOUBLE PRECISION,
    light DOUBLE PRECISION
);

SELECT create_hypertable('sensor_data', 'time');
```

## License

MIT