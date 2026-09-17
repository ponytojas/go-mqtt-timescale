# MQTT to Supabase Connector

This service subscribes to MQTT sensor messages and stores them in the existing
`labtools.sensor_data` table through Supabase's Data REST API, using the
community `supabase-go` client. It does not open a PostgreSQL connection or
require TimescaleDB.

## Database table

The service expects this table to have the columns shown below. It never creates
or changes database objects.

```sql
CREATE TABLE labtools.sensor_data (
  time timestamptz PRIMARY KEY,
  temperature double precision,
  humidity double precision,
  light double precision,
  device_id text NOT NULL
);
```

## Configuration

In Supabase, open **Connect** to get the Project URL and **Settings → API Keys**
to get a secret key. Configure those as `SUPABASE_URL` and `SUPABASE_API_KEY`.
The service calls `POST /rest/v1/sensor_data` with `Content-Profile: labtools`.

Before deploying, add `labtools` to the Data API's exposed schemas and grant the
API role insert access to `labtools.sensor_data`. If Row Level Security is enabled,
create an insert policy for the key's role. A server process should use a secret
key and keep it out of source control.

```sh
export SUPABASE_URL='https://PROJECT_REF.supabase.co'
export SUPABASE_API_KEY='sb_secret_...'
export SUPABASE_SCHEMA='labtools'
export SUPABASE_TABLE_NAME='sensor_data'
export MQTT_BROKER='ssl://mqtt.example.com:8883'
export MQTT_TOPIC='sensors/data'
export MQTT_USERNAME='username'
export MQTT_PASSWORD='password'
go run ./cmd
```

`SUPABASE_SCHEMA` defaults to `labtools` and `SUPABASE_TABLE_NAME` defaults to
`sensor_data`.

The complete MQTT configuration is:

| Environment variable | Default |
| --- | --- |
| `MQTT_BROKER` (or legacy `MQTT_BROKER_URL`) | `https://mqtt.ponytojas.dev` |
| `MQTT_PORT` | `8883` |
| `MQTT_CLIENT_ID` | `go-mqtt-client` |
| `MQTT_TOPIC` | `sensor/#` |
| `MQTT_USERNAME` | empty |
| `MQTT_PASSWORD` | empty |

## Message format

```json
{
  "timestamp": "2026-09-17T15:04:05Z",
  "temperature": 24.5,
  "humidity": 65.2,
  "light": 850,
  "device_id": "sensor-01"
}
```

`device_id` is required. When `timestamp` is omitted, the service records its
current time. A `light` value of zero is valid and is stored.

## Supabase integration test

The integration test inserts one uniquely labelled row, then deletes that exact
row. It is skipped unless explicitly enabled.

```sh
set -a
source .env
set +a
SUPABASE_INTEGRATION_TEST=true go test ./internal/supabase -run TestSupabaseInsertAndDelete -v
```

Add an existing `labtools.rooms.sensor_id` value to `.env` before running it:

```env
SUPABASE_TEST_DEVICE_ID=existing-sensor-id
```

`sensor_data.device_id` has a foreign key, so the test uses this real sensor ID
while inserting and deleting only its temporary reading. The configured key needs
both `INSERT` and `DELETE` permission under the `labtools.sensor_data` RLS policies.

## Docker

```sh
docker build -t mqtt-supabase .
docker run -d --name mqtt-supabase \
  -e SUPABASE_URL='https://PROJECT_REF.supabase.co' \
  -e SUPABASE_API_KEY='sb_secret_...' \
  -e MQTT_BROKER='ssl://mqtt.example.com:8883' \
  -e MQTT_TOPIC='sensors/data' \
  mqtt-supabase
```
