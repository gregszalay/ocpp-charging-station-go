# OCPP Charging Station Frontend

This web application serves as a local UI for a charging station.

The application is designed to run in kiosk mode on a touch display connected to a RaspberryPi 4b or similar hardware.

## Main Features:

- List EVSEs
- Show basic status information about each EVSE
- Start and stop charging with RFID authentication

## The backend

The app is designed to connect to the local charging station backend http server. (See [ocpp-charging-station-go](https://github.com/gregszalay/ocpp-charging-station-go) for details.)

## Build & Serve

    npm run build
    serve -s build -l 3001
