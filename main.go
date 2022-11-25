package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gregszalay/ocpp-charging-station-go/chargingstation"
)

var ocpp_host = flag.String("h", "localhost:3000", "ocpp websocket server host")
var ocpp_url = flag.String("u", "/ocpp", "ocpp URL")
var ocpp_station_id = flag.String("id", "CS001", "id of the charging station")

var simulationinprogress bool = false

func main() {

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	csms_url := url.URL{Scheme: "ws", Host: *ocpp_host, Path: *ocpp_url + "/" + *ocpp_station_id}
	fmt.Printf("connecting to CSMS through URL: %s\n", csms_url.String())

	chargingstation.ChargingStation(csms_url)
	
	return

	
}
