package chargingstation

import (
	"errors"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/gregszalay/ocpp-charging-station-go/displayserver"
	"github.com/gregszalay/ocpp-charging-station-go/evsemanager"
	"github.com/gregszalay/ocpp-charging-station-go/ocppclient"
	"github.com/gregszalay/ocpp-charging-station-go/transactions"
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
	EVSEIdsToTxsMap     map[int]*transactions.Transaction
}

var cs_new *ChargingStation

func CreateAndRunChargingStation(_csms_url url.URL, evseIPs []string) (*ChargingStation, error) {

	// Create new Charging Station
	cs_new = &ChargingStation{
		Csms_url:            _csms_url, //TODO more than one csms?
		Evses:               make(map[int]*evsemanager.EVSE),
		OcppClient:          nil,
		UI_callbacks:        nil,
		SetVariablesHandler: func(call wrappers.CALL) {},
		EVSEIdsToTxsMap:     make(map[int]*transactions.Transaction),
	}

	// Connect EVSEs
	if len(evseIPs) == 0 {
		return nil, errors.New("failed to create CS, at least 1 EVSE must be provided")
	}
	for i, ip := range evseIPs {
		// TODO check evse numbering standard
		if evse, err := evsemanager.CreateAndRunEVSE(i, ip); err != nil {
			log.Error("Unable to create EVSE instance", err)
			return nil, err
		} else {
			cs_new.Evses[i] = evse
		}
	}

	// Connect OCPP Client
	if ocpp_cl, err := ocppclient.CreateAndRunOCPPClient(_csms_url); err != nil {
		log.Error("failed to create OCPP client")
		return nil, err
	} else {
		cs_new.OcppClient = ocpp_cl
	}

	cs_new.SendBootNotification()

	for _, evse := range cs_new.Evses {
		cs_new.SendStatusNotification(evse)
	}

	time.Sleep(time.Millisecond * 3000)

	// Set up the UI logic
	cs_new.UI_callbacks = &displayserver.UICallbacks{
		OnStartButtonPress: func(evseId int, rfid string) {
			evse := cs_new.Evses[evseId]
			evse.EnableCharging()
			new_tx, err := cs_new.StartTransaction(evse, rfid)
			if err != nil {
				log.Error("Unable to start new transaction")
			}
			cs_new.EVSEIdsToTxsMap[evse.Id] = new_tx
		},
		OnStopButtonPress: func(evseId int, rfid string) {
			evse := cs_new.Evses[evseId]
			evse.DisableCharging()
			tx := cs_new.EVSEIdsToTxsMap[evse.Id]
			cs_new.EndTransaction(evse, tx, rfid)
		},
		OnGetChargeStatus: func(evseId int) displayserver.EVSEStatusDataForUI {
			evse := cs_new.Evses[evseId]
			data := displayserver.EVSEStatusDataForUI{
				IsEVConnected:              evse.IsEVConnected,
				IsChargingEnabled:          evse.IsChargingEnabled,
				IsCharging:                 evse.IsCharging,
				IsError:                    evse.IsError,
				EnergyActiveNet_kwh_float:  float64(evse.EnergyActiveNet_wh) / 1000,
				PowerActiveImport_kw_float: float64(evse.PowerActiveImport_w) / 1000,
			}
			return data
		},
		OnGetEVSEsActiveIds: func() []int {
			evses := make([]int, 0)
			evseNumber := 0
			for evseId, _ := range cs_new.Evses {
				evses = append(evses, evseId)
				evseNumber++
			}
			return evses
		},
	}

	go displayserver.Start(*cs_new.UI_callbacks)

	log.Info("got here ")

	//HEARTBEAT JOB
	go func() {
		ticker_status := time.NewTicker(time.Second * 10)
		defer ticker_status.Stop()
		for {
			select {
			case t := <-ticker_status.C:
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
				log.Info("Sending heartbeat")
				cs_new.OcppClient.Send(ocppclient.AsyncOcppCall{
					Message:         *call_wrapper,
					SuccessCallback: func(result wrappers.CALLRESULT) { log.Info("Heartbeat message sent") },
					ErrorCallback:   func(result wrappers.CALLERROR) { log.Info("Heartbeat message not sent") },
				})
			default:
			}
		}
	}()

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
