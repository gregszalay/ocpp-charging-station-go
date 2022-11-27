# ocpp-charging-station-go

Simulates OCPP 2.0.1 client connections

## Features:

    Simulates all the basic messages of a bootup and a charging event

    (boot->status->start->auth->update->stop->status)

## Physical topology

    CSMS <==JSON/WS/TCP/internet==> CS <==TCP/LAN==> EVSE

## Quick Start

1.  Build

        go build main.go

2.  Run (example)

        ./main -debugl info -host localhost:3000 -url /ocpp -id CS123 192.168.1.71:80 192.168.1.65:80

    Options:

        -host: the host where the csms is running
        -url: the URL endpoint
        -id: the ID of the charging station
        -list of IP adresses of the EVSE servers on the LAN
