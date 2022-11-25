package chargingstation

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gregszalay/ocpp-charging-station-go/evsemanager"
	"github.com/gregszalay/ocpp-charging-station-go/ocppclient"
	"github.com/gregszalay/ocpp-messages-go/types/TransactionEventRequest"
	tx_lib "github.com/gregszalay/ocpp-messages-go/types/TransactionEventRequest"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
	log "github.com/sirupsen/logrus"
)

type Transaction struct {
	Id           string
	TxSeqNo      int
	IsInProgress bool
	evse         evsemanager.EVSE
}

// Starting transaction - E02 - Cable Plugin First
func (cs *ChargingStation) StartNewTx(evse evsemanager.EVSE) (*Transaction, error) {

	new_tx := &Transaction{
		Id:           uuid.New().String(), // Generate new UUID for the transaction
		TxSeqNo:      0,
		IsInProgress: false,
		evse:         evse,
	}

	// Send StatusNotificationRequest
	cs.SendStatusNotification(new_tx.evse)

	// ==> TXEventReq: Started, CablePluggedIn
	new_tx.sendTxEventReq(
		TransactionEventRequest.TransactionEventEnumType_1_Started,
		new_tx.evse.EnergyActiveNet_kwh_times100,
		new_tx.evse.PowerActiveImport_kw_times100,
		TransactionEventRequest.TriggerReasonEnumType_1_CablePluggedIn,
		new_tx.Id,
		func() {},
		func() { log.Error("Cannot continue with TX") }, //TODO stop charge process
	)

	// Read RFID string from std input. Send AuthorizeRequest to CSMS
	cs.authorizeWithRFID(func() {

		// ==> TXEventReq: Updated, Authorized
		new_tx.sendTxEventReq(
			TransactionEventRequest.TransactionEventEnumType_1_Updated,
			new_tx.evse.EnergyActiveNet_kwh_times100,
			new_tx.evse.PowerActiveImport_kw_times100,
			TransactionEventRequest.TriggerReasonEnumType_1_Authorized,
			new_tx.Id,
			func() {},
			func() { log.Error("Cannot continue with TX") }, //TODO stop charge process
		)
		fmt.Println("Starting TX updates loop")
		new_tx.IsInProgress = true
		go new_tx.loop()
	}, func() {
		log.Error("Authorization failed")
	})

	return new_tx, nil

}

func (tx *Transaction) loop() {
	for {

		if tx.IsInProgress == false {
			break
		}

		// ==> TXEventReq: Updated, ChargingStateChanged
		tx.sendTxEventReq(
			TransactionEventRequest.TransactionEventEnumType_1_Updated,
			tx.evse.EnergyActiveNet_kwh_times100,
			tx.evse.PowerActiveImport_kw_times100,
			TransactionEventRequest.TriggerReasonEnumType_1_ChargingStateChanged,
			tx.Id,
			func() {},
			func() { log.Error("Cannot continue with TX") }, //TODO stop charge process
		)

		time.Sleep(time.Second * 5)

	}
}

func (tx *Transaction) EndTx() {
	// // ==> AuthorizeReq. Read RFID string from std input. Send AuthorizeRequest to CSMS
	authorizeWithRFID(
		func() {
			// ==> TXEventReq: Updated, StopAuthorized. Notify the CSMS that the driver is authorized to stop the Transaction
			tx.sendTxEventReq(
				TransactionEventRequest.TransactionEventEnumType_1_Updated,
				tx.evse.EnergyActiveNet_kwh_times100,
				tx.evse.PowerActiveImport_kw_times100,
				TransactionEventRequest.TriggerReasonEnumType_1_StopAuthorized,
				tx.Id,
				func() {
					// Set the callback to fire when the EV plug disconnected by the driver
					tx.evse.OnEVDisconnected_fire_once = func() {
						// ==> TXEventReq: Updated, StopAuthorized. Notify the CSMS that the Transaction has ended
						tx.sendTxEventReq(
							TransactionEventRequest.TransactionEventEnumType_1_Ended,
							tx.evse.EnergyActiveNet_kwh_times100,
							tx.evse.PowerActiveImport_kw_times100,
							TransactionEventRequest.TriggerReasonEnumType_1_EVCommunicationLost,
							tx.Id,
							func() {
								// ==> SendStatusNotificationReq. Notify the CSMS that the EVSE is available again
								SendStatusNotification(tx.evse)
								tx.IsInProgress = false
							},
							func() { log.Error("==> TXEventReq: Updated, StopAuthorized FAILED") })
					}
				},
				func() { log.Error("Cannot continue with TX") }, //TODO stop charge process
			)
		},
		func() {
			log.Error("Authorization failed")
		})
}

func (tx *Transaction) sendTxEventReq(
	_eventType tx_lib.TransactionEventEnumType_1,
	_EnergyActiveNet_kwh_times100 int,
	_PowerActiveImport_kw_times100 int,
	_triggerR tx_lib.TriggerReasonEnumType_1,
	_transactionId string,
	onSuccess func(),
	onFailure func(),
) {
	energy_active_net_type := tx_lib.MeasurandEnumType_1_EnergyActiveNet
	power_active_import_type := tx_lib.MeasurandEnumType_1_PowerActiveImport
	tx_req := &tx_lib.TransactionEventRequestJson{
		EventType: _eventType,
		MeterValue: []tx_lib.MeterValueType{
			tx_lib.MeterValueType{
				SampledValue: []tx_lib.SampledValueType{
					tx_lib.SampledValueType{
						Measurand: &energy_active_net_type,
						UnitOfMeasure: &tx_lib.UnitOfMeasureType{
							Multiplier: 3,
							Unit:       "Wh",
						},
						Value: float64(_EnergyActiveNet_kwh_times100) / 100,
					},
					tx_lib.SampledValueType{
						Measurand: &power_active_import_type,
						UnitOfMeasure: &tx_lib.UnitOfMeasureType{
							Multiplier: 3,
							Unit:       "W",
						},
						Value: float64(_PowerActiveImport_kw_times100) / 100,
					},
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
		SeqNo:     tx.TxSeqNo,
		Timestamp: time.Now().Format(time.RFC3339),
		TransactionInfo: tx_lib.TransactionType{
			TransactionId: _transactionId,
		},
		TriggerReason: _triggerR,
	}

	tx.TxSeqNo += 1

	call_wrapper := &wrappers.CALL{
		MessageTypeId: wrappers.CALL_TYPE,
		MessageId:     uuid.New().String(),
		Action:        "TransactionEvent",
		Payload:       *tx_req,
	}

	ocppclient.Send(ocppclient.AsyncOcppCall{
		Message: *call_wrapper,
		SuccessCallback: func(callresult wrappers.CALLRESULT) {
			fmt.Println("TransactionEventReq received by CSMS")
			onSuccess()
		},
		ErrorCallback: func(wrappers.CALLERROR) {
			log.Error("TransactionEventReq NOT received by CSMS")
			onFailure()
		},
	})

}
