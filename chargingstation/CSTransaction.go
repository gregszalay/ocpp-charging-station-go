package chargingstation

import (
	"fmt"
	"time"

	"github.com/gregszalay/ocpp-charging-station-go/evsemanager"
	"github.com/gregszalay/ocpp-charging-station-go/ocppclient"
	"github.com/gregszalay/ocpp-charging-station-go/transactions"
	"github.com/gregszalay/ocpp-messages-go/types/TransactionEventRequest"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
	log "github.com/sirupsen/logrus"
)

// Starting transaction - E02 - Cable Plugin First
func (cs *ChargingStation) StartTransaction(evse *evsemanager.EVSE) (*transactions.Transaction, error) {

	tx_new, _ := transactions.CreateTransaction(*evse)

	// Send StatusNotificationRequest
	cs.SendStatusNotification(evse)

	// ==> TXEventReq: Started, CablePluggedIn
	tx_event_req, _ := tx_new.MakeTransactionEventReq(
		TransactionEventRequest.TransactionEventEnumType_1_Started,
		TransactionEventRequest.TriggerReasonEnumType_1_CablePluggedIn,
	)
	cs.OcppClient.Send(ocppclient.AsyncOcppCall{
		Message: tx_event_req,
		SuccessCallback: func(callresult wrappers.CALLRESULT) {
			fmt.Println("TransactionEventReq received by CSMS")
		},
		ErrorCallback: func(wrappers.CALLERROR) {
			log.Error("TransactionEventReq NOT received by CSMS")
		},
	})

	// Read RFID string from std input. Send AuthorizeRequest to CSMS
	cs.authorizeWithRFID(func() {
		// ==> TXEventReq: Updated, Authorized
		tx_event_req, _ := tx_new.MakeTransactionEventReq(
			TransactionEventRequest.TransactionEventEnumType_1_Updated,
			TransactionEventRequest.TriggerReasonEnumType_1_Authorized,
		)
		cs.OcppClient.Send(ocppclient.AsyncOcppCall{
			Message: tx_event_req,
			SuccessCallback: func(callresult wrappers.CALLRESULT) {
				fmt.Println("TransactionEventReq received by CSMS")
			},
			ErrorCallback: func(wrappers.CALLERROR) {
				log.Error("TransactionEventReq NOT received by CSMS")
			}})

		go func() {
			for {
				if tx_new.IsInProgress == false {
					break
				}
				// ==> TXEventReq: Updated, ChargingStateChanged
				tx_event_req, _ := tx_new.MakeTransactionEventReq(
					TransactionEventRequest.TransactionEventEnumType_1_Updated,
					TransactionEventRequest.TriggerReasonEnumType_1_ChargingStateChanged,
				)
				cs.OcppClient.Send(ocppclient.AsyncOcppCall{
					Message: tx_event_req,
					SuccessCallback: func(callresult wrappers.CALLRESULT) {
						fmt.Println("TransactionEventReq received by CSMS")
					},
					ErrorCallback: func(wrappers.CALLERROR) {
						log.Error("TransactionEventReq NOT received by CSMS")
					}})
				time.Sleep(time.Second * 5)
			}
		}()
	}, func() {
		log.Error("Authorization failed")
	})
	return tx_new, nil
}

func (cs *ChargingStation) EndTransaction(evse *evsemanager.EVSE, tx *transactions.Transaction) {

	// ==> AuthorizeReq. Read RFID string from std input. Send AuthorizeRequest to CSMS
	cs.authorizeWithRFID(
		func() { // Auth success:
			// ==> TXEventReq: Updated, StopAuthorized. Notify the CSMS that the driver is authorized to stop the Transaction
			tx_event_req, _ := tx.MakeTransactionEventReq(
				TransactionEventRequest.TransactionEventEnumType_1_Updated,
				TransactionEventRequest.TriggerReasonEnumType_1_StopAuthorized,
			)
			cs.OcppClient.Send(ocppclient.AsyncOcppCall{
				Message: tx_event_req,
				SuccessCallback: func(callresult wrappers.CALLRESULT) {
					fmt.Println("TransactionEventReq received by CSMS")
					// Set the callback to fire when the EV plug disconnected by the driver
					evse.OnEVDisconnected_fire_once = func() {
						// ==> TXEventReq: Updated, StopAuthorized. Notify the CSMS that the Transaction has ended
						tx_event_req, _ := tx.MakeTransactionEventReq(
							TransactionEventRequest.TransactionEventEnumType_1_Ended,
							TransactionEventRequest.TriggerReasonEnumType_1_EVCommunicationLost,
						)
						cs.OcppClient.Send(ocppclient.AsyncOcppCall{
							Message: tx_event_req,
							SuccessCallback: func(callresult wrappers.CALLRESULT) {
								fmt.Println("TransactionEventReq received by CSMS")
								// ==> SendStatusNotificationReq. Notify the CSMS that the EVSE is available again
								cs.SendStatusNotification(evse)
								tx.IsInProgress = false
							},
							ErrorCallback: func(wrappers.CALLERROR) {
								log.Error("TransactionEventReq NOT received by CSMS")
							}})
					}
				},
				ErrorCallback: func(wrappers.CALLERROR) {
					log.Error("TransactionEventReq NOT received by CSMS")
				}})
		},
		func() { // Auth failed:
			log.Error("Authorization failed")
		})
}
