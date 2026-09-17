#!/bin/bash

# This script publishes test data to the MQTT broker

# Default values
BROKER="${MQTT_BROKER:-mqtt.ponytojas.dev}"
PORT="${MQTT_PORT:-8883}"
TOPIC="sensors/data"
COUNT=10
DELAY=5
DEVICE_ID="${DEVICE_ID:-test-sensor}"
USERNAME="${MQTT_USERNAME:-}"
PASSWORD="${MQTT_PASSWORD:-}"
USE_TLS=true

# Help function
function show_help {
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  -b, --broker BROKER     MQTT broker address (default: mqtt.ponytojas.dev)"
    echo "  -p, --port PORT         MQTT broker port (default: 8883)"
    echo "  -t, --topic TOPIC       MQTT topic to publish to (default: sensors/data)"
    echo "  -c, --count COUNT       Number of messages to publish (default: 10)"
    echo "  -d, --delay DELAY       Delay between messages in seconds (default: 5)"
		echo "  -i, --device-id ID      Device identifier (default: test-sensor)"
    echo "  -u, --username USER     MQTT username"
    echo "  -P, --password PASS     MQTT password"
    echo "  --no-tls                Disable TLS (use plain TCP)"
    echo "  -h, --help              Show this help message"
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -b|--broker)
            BROKER="$2"
            shift; shift ;;
        -p|--port)
            PORT="$2"
            shift; shift ;;
        -t|--topic)
            TOPIC="$2"
            shift; shift ;;
        -c|--count)
            COUNT="$2"
            shift; shift ;;
        -d|--delay)
            DELAY="$2"
            shift; shift ;;
		-i|--device-id)
			DEVICE_ID="$2"
			shift; shift ;;
        -u|--username)
            USERNAME="$2"
            shift; shift ;;
        -P|--password)
            PASSWORD="$2"
            shift; shift ;;
        --no-tls)
            USE_TLS=false
            shift ;;
        -h|--help)
            show_help ;;
        *)
            echo "Unknown option: $1"
            show_help ;;
    esac
done

# Check if mosquitto_pub is installed
if ! command -v mosquitto_pub &> /dev/null; then
    echo "Error: mosquitto_pub could not be found"
    echo "Please install mosquitto-clients package"
    exit 1
fi

# Prepare common options
PUB_CMD=(mosquitto_pub -h "$BROKER" -p "$PORT" -t "$TOPIC")

if [ -n "$USERNAME" ]; then
    PUB_CMD+=(-u "$USERNAME" -P "$PASSWORD")
fi

if [ "$USE_TLS" = true ]; then
    PUB_CMD+=(--capath /etc/ssl/certs --tls-version tlsv1.2)
fi

# Publish test data
echo "Publishing $COUNT test messages to $BROKER:$PORT on topic $TOPIC"
echo "Press Ctrl+C to stop"

for ((i=1; i<=$COUNT; i++)); do
    TEMP=$(echo "scale=1; 20 + $(( RANDOM % 15 )).$((RANDOM % 10))" | bc)
    HUM=$(echo "scale=1; 30 + $(( RANDOM % 70 )).$((RANDOM % 10))" | bc)
    LIGHT=$(echo "scale=0; 100 + $(( RANDOM % 900 ))" | bc)
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    JSON="{\"timestamp\":\"$TIMESTAMP\",\"temperature\":$TEMP,\"humidity\":$HUM,\"light\":$LIGHT,\"device_id\":\"$DEVICE_ID\"}"

    echo "Publishing message $i/$COUNT: $JSON"
    "${PUB_CMD[@]}" -m "$JSON"

    if [ $i -lt $COUNT ]; then
        echo "Waiting $DELAY seconds before next message..."
        sleep $DELAY
    fi
done

echo "Done!"
