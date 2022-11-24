package chargingstation

import (
	"fmt"
	"net/url"

	"github.com/gregszalay/ocpp-charging-station-go/evse"
	"github.com/gregszalay/ocpp-charging-station-go/ocppclient"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
)

var incoming_call_handlers map[string]func(wrappers.CALL) = map[string]func(wrappers.CALL){
	"SetVariables": func(callresult wrappers.CALL) {
		fmt.Printf("Callback invoked for SetVariables\n")
	},
}

func ChargingStation(u url.URL) {
	fmt.Println("Initializing chargingstation...")
	ocppclient.Connect(u, incoming_call_handlers)
	evse.Connect("192.168.1.71:80")
}
