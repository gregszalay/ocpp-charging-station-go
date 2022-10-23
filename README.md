# ocpp-charging-station-simulator

Simulates OCPP 2.0.1 client connections

## Features:

    Simulates all the basic messages of a bootup and a charging event

    (boot->status->start->auth->update->stop->status)

## Quick Start

1.  Build

        go build main.go

2.  Run (example)

        ./main -h localhost:3000 -u /ocpp -id CS123

    Options:

        -h: the host where the csms is running
        -u: the URL endpoint
        -id: the is of the charging station you are simulating
