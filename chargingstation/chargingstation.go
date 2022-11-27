package chargingstation

import (
	"net/url"

	"github.com/google/uuid"
	"github.com/gregszalay/ocpp-charging-station-go/displayserver"
	"github.com/gregszalay/ocpp-charging-station-go/evsemanager"
	"github.com/gregszalay/ocpp-charging-station-go/ocppclient"
	"github.com/gregszalay/ocpp-charging-station-go/transactions"
	"github.com/gregszalay/ocpp-messages-go/types/BootNotificationRequest"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
	log "github.com/sirupsen/logrus"
)

type ChargingStation struct {
	Csms_url            url.URL
	Evses               map[int]*evsemanager.EVSE
	OcppClient          *ocppclient.OCPPClient
	UI_callbacks        *displayserver.UICallbacks
	SetVariablesHandler func(wrappers.CALL)
	EVSEIdsToTxsMap     map[int]*transactions.Transaction
}

func CreateAndRunChargingStation(_csms_url url.URL, evseIPs []string) (*ChargingStation, error) {

	log.Info("1")

	// Create new Charging Station
	cs_new := &ChargingStation{
		Csms_url:            _csms_url, //TODO more than one csms?
		Evses:               make(map[int]*evsemanager.EVSE),
		OcppClient:          nil,
		UI_callbacks:        nil,
		SetVariablesHandler: func(call wrappers.CALL) {},
		EVSEIdsToTxsMap:     make(map[int]*transactions.Transaction),
	}

	// Connect EVSEs
	// if len(evseIPs) == 0 {
	// 	return nil, errors.New("failed to create CS, at least 1 EVSE must be provided")
	// }
	// for i, ip := range evseIPs {
	// 	// TODO check evse numbering standard
	// 	if evse, err := evsemanager.CreateAndRunEVSE(i, ip); err != nil {
	// 		log.Error("Unable to create EVSE instance", err)
	// 		return nil, err
	// 	} else {
	// 		cs_new.Evses[i] = evse
	// 	}
	// }

	// Connect OCPP Client
	if ocpp_cl, err := ocppclient.CreateAndRunOCPPClient(_csms_url); err != nil {
		log.Error("failed to create OCPP client")
		return nil, err
	} else {
		cs_new.OcppClient = ocpp_cl
	}

	cs_new.SendBootNotification()

	// Set up the UI logic
	// cs_new.UI_callbacks = &displayserver.UICallbacks{
	// 	OnStartButtonPress: func(evseId int) {
	// 		evse := *cs_new.Evses[evseId]
	// 		new_tx, err := cs_new.StartTransaction(&evse)
	// 		if err != nil {
	// 			log.Error("Unable to start new transaction")
	// 		}
	// 		cs_new.EVSEIdsToTxsMap[evse.Id] = new_tx
	// 	},
	// 	OnStopButtonPress: func(evseId int) {
	// 		evse := *cs_new.Evses[evseId]
	// 		tx := cs_new.EVSEIdsToTxsMap[evse.Id]
	// 		cs_new.EndTransaction(&evse, tx)
	// 	},
	// 	OnGetChargeStatus: func(evseId int) string {
	// 		evse := *cs_new.Evses[evseId]
	// 		data := displayserver.EVSEStatusDataForUI{
	// 			EnergyActiveNet_kwh_float:  float64(evse.EnergyActiveNet_kwh_times100) / 100,
	// 			PowerActiveImport_kw_float: float64(evse.PowerActiveImport_kw_times100) / 100,
	// 		}
	// 		json_str, err := json.Marshal(data)
	// 		if err != nil {
	// 			log.Error("Failed to marshal UI data")
	// 		}
	// 		return string(json_str)
	// 	},
	// 	OnGetEVSEsActiveIds: func() string {
	// 		evses := make([]int, 1)
	// 		evseNumber := 0
	// 		for evseId, _ := range cs_new.Evses {
	// 			evses = append(evses, evseId)
	// 			evseNumber++
	// 		}
	// 		json_str, err := json.Marshal(evses)
	// 		if err != nil {
	// 			log.Error("Failed to marshal UI data")
	// 		}
	// 		return string(json_str)
	// 	},
	// }

	// go func() {
	// 	displayserver.Start(*cs_new.UI_callbacks)
	// }()

	// HEARTBEAT JOB
	// ticker_status := time.NewTicker(time.Second)
	// go func() {
	// 	defer ticker_status.Stop()
	// 	for {
	// 		select {
	// 		case t := <-ticker_status.C:
	// 			_ = t
	// 			heartbeatRequest := &HeartbeatRequest.HeartbeatRequestJson{
	// 				CustomData: &HeartbeatRequest.CustomDataType{
	// 					VendorId: "example-station-vendor",
	// 				},
	// 			}
	// 			call_wrapper := &wrappers.CALL{
	// 				MessageTypeId: wrappers.CALL_TYPE,
	// 				MessageId:     uuid.New().String(),
	// 				Action:        "Heartbeat",
	// 				Payload:       *heartbeatRequest,
	// 			}
	// 			cs_new.OcppClient.Send(ocppclient.AsyncOcppCall{
	// 				Message:         *call_wrapper,
	// 				SuccessCallback: func(result wrappers.CALLRESULT) {},
	// 				ErrorCallback:   func(result wrappers.CALLERROR) {},
	// 			})
	// 		default:
	// 		}
	// 	}
	// }()

	go func() {

		for ocpp_call_from_CSMS := range cs_new.OcppClient.Calls_received {
			switch ocpp_call_from_CSMS.Action {
			case "SetVariables":
				log.Info("handler for SetVariables called")
				//cs_new.OcppClient.SetVariablesHandler(ocpp_call_from_CSMS) // TODO setvariableshandler type as param
			default:
				log.Warning("No handler found for this CSMS request")
			}
		}
	}()

	return cs_new, nil
}

func (cs *ChargingStation) ShutDown() {
	//TODO
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
