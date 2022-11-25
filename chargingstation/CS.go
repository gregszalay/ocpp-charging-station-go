package chargingstation

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/gregszalay/ocpp-charging-station-go/displayserver"
	"github.com/gregszalay/ocpp-charging-station-go/evsemanager"
	"github.com/gregszalay/ocpp-charging-station-go/ocppclient"
	"github.com/gregszalay/ocpp-messages-go/types/BootNotificationRequest"
	"github.com/gregszalay/ocpp-messages-go/types/HeartbeatRequest"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
	log "github.com/sirupsen/logrus"
)

type ChargingStation struct {
	Csms_url            url.URL
	Evses               map[int]*evsemanager.EVSE
	OcppClient          *ocppclient.OCPPClient
	UI_callbacks        *displayserver.UICallbacks
	SetVariablesHandler func(wrappers.CALL)
	Transactions        map[string]Transaction
}

func CreateChargingStation(_csms_url url.URL) (*ChargingStation, error) {
	log.Info("Creating charging station...")

	// Create new Charging Station
	cs_new := &ChargingStation{
		Csms_url:            _csms_url,
		Evses:               make(map[int]*evsemanager.EVSE),
		OcppClient:          nil,
		UI_callbacks:        nil,
		SetVariablesHandler: func(call wrappers.CALL) {},
	}

	// Create new OCPP Client
	if ocpp_cl, err := ocppclient.CreateOCPPClient(); err != nil {
		log.Error("failed to create OCPP client")
		return nil, err
	} else {
		cs_new.OcppClient = ocpp_cl
	}

	// Create EVSE #1 instance and connect to its tcp server
	evse_1_id := 1
	if evse_1, err := evsemanager.ConnectNewEVSE(evse_1_id, "192.168.1.71:80"); err != nil {
		log.Error("Unable to create EVSE instance", err)
		return nil, err
	} else {
		cs_new.Evses[evse_1_id] = evse_1
	}

	// Create more EVSEs if you like...

	// Set up the UI logic
	cs_new.UI_callbacks = &displayserver.UICallbacks{
		OnStartButtonPress: func(evseId int) {
			evse := *cs_new.Evses[evseId]
			new_tx, err := cs_new.StartNewTx(evse)
			if err != nil {
				log.Error("Unable to start new transaction")
			}
			evse.LatestTransactionId = new_tx.Id
		},
		OnStopButtonPress: func(evseId int) {
			evse := *cs_new.Evses[evseId]
			tx := cs_new.Transactions[evse.LatestTransactionId]
			tx.EndTx()
		},
		OnGetChargeStatus: func(evseId int) string {
			evse := *cs_new.Evses[evseId]
			data := displayserver.EVSEStatusDataForUI{
				EnergyActiveNet_kwh_float:  float64(evse.EnergyActiveNet_kwh_times100) / 100,
				PowerActiveImport_kw_float: float64(evse.PowerActiveImport_kw_times100) / 100,
			}
			json_str, err := json.Marshal(data)
			if err != nil {
				log.Error("Failed to marshal UI data")
			}
			return string(json_str)
		},
	}
	return cs_new, nil
}

func (cs *ChargingStation) BootUp() {

	// Create Charging Station and connect it to the CSMS in the cloud
	cs.OcppClient.ConnectToCSMS(cs.Csms_url)

	displayserver.Start(*cs.UI_callbacks)

	cs.SendBootNotification()

	// Start handling incoming calls from CSMS
	go func() {
		for ocpp_call_from_CSMS := range cs.OcppClient.Calls_received {
			cs.handleCALLFromCSMS(ocpp_call_from_CSMS)
		}
	}()

	// Heartbeat loop
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case t := <-ticker.C:
				_ = t
				heartbeatRequest := &HeartbeatRequest.HeartbeatRequestJson{
					CustomData: &HeartbeatRequest.CustomDataType{
						VendorId: "example-station-vendor",
					},
				}
				call_wrapper := &wrappers.CALL{
					MessageTypeId: wrappers.CALL_TYPE,
					MessageId:     uuid.New().String(),
					Action:        "Heartbeat",
					Payload:       *heartbeatRequest,
				}
				cs.OcppClient.Send(ocppclient.AsyncOcppCall{
					Message:         *call_wrapper,
					SuccessCallback: func(result wrappers.CALLRESULT) {},
					ErrorCallback:   func(result wrappers.CALLERROR) {},
				})
			default:
			}
		}
	}()

}

func (cs *ChargingStation) handleCALLFromCSMS(call wrappers.CALL) {
	switch call.Action {
	case "SetVariables":
		log.Info("handler for SetVariables called")
		cs.SetVariablesHandler(call) // TODO setvariableshandler type as param
	default:
		log.Warning("No handler found for this CSMS request")
	}

}

func (cs *ChargingStation) SendBootNotification() {

	bootNotificationRequest := BootNotificationRequest.BootNotificationRequestJson{
		Reason: "PowerUp",
		ChargingStation: BootNotificationRequest.ChargingStationType{
			Model:      "super-charger-6000",
			VendorName: "WattsUp",
		},
	}
	call_wrapper := &wrappers.CALL{
		MessageTypeId: wrappers.CALL_TYPE,
		MessageId:     uuid.New().String(),
		Action:        "BootNotification",
		Payload:       bootNotificationRequest,
	}

	cs.OcppClient.Send(ocppclient.AsyncOcppCall{
		Message:         *call_wrapper,
		SuccessCallback: func(result wrappers.CALLRESULT) {},
		ErrorCallback:   func(result wrappers.CALLERROR) {},
	})
}
