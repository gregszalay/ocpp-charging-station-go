package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"

	"github.com/gregszalay/ocpp-charging-station-go/chargingstation"
	log "github.com/sirupsen/logrus"
)

var debug_level = flag.String("debugl", "Info", "Debug log level")
var ocpp_host = flag.String("host", "localhost:3000", "ocpp websocket server host")
var ocpp_url = flag.String("url", "/ocpp", "ocpp URL")
var ocpp_station_id = flag.String("id", "CS001", "id of the charging station")

func main() {
	setLogLevel(*debug_level)

	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	csms_url := url.URL{Scheme: "wss", Host: *ocpp_host, Path: *ocpp_url + "/" + *ocpp_station_id}
	fmt.Printf("connecting to CSMS through URL: %s\n", csms_url.String())

	evseIPs := flag.Args() // e.g. "192.168.1.71:80"

	_, err := chargingstation.CreateAndRunChargingStation(csms_url, evseIPs)
	if err != nil {
		log.Error("failed to create charging station: ", err)
		return
	}

	// FOR TESTING ONLY
	//displaytest.RunDisplayTest()

	func() {
		for {
			select {
			case <-interrupt:
				os.Exit(1)
			default:
			}
		}
	}()

}

func setLogLevel(levelName string) {
	switch levelName {
	case "Panic":
		log.SetLevel(log.PanicLevel)
	case "Fatal":
		log.SetLevel(log.FatalLevel)
	case "Error":
		log.SetLevel(log.ErrorLevel)
	case "Warn":
		log.SetLevel(log.WarnLevel)
	case "Info":
		log.SetLevel(log.InfoLevel)
	case "Debug":
		log.SetLevel(log.DebugLevel)
	case "Trace":
		log.SetLevel(log.TraceLevel)
	}
}
