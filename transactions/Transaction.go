package transactions

import (
	"time"

	"github.com/google/uuid"
	"github.com/gregszalay/ocpp-charging-station-go/evsemanager"
	tx_lib "github.com/gregszalay/ocpp-messages-go/types/TransactionEventRequest"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
)

type Transaction struct {
	Id           string
	Evse         evsemanager.EVSE
	TxSeqNo      int
	IsInProgress bool
}

func CreateTransaction(evse evsemanager.EVSE) (*Transaction, error) {

	tx_new := &Transaction{
		Id:           uuid.New().String(), // Generate new UUID for the transaction
		Evse:         evse,
		TxSeqNo:      0,
		IsInProgress: false,
	}

	return tx_new, nil
}

func (tx *Transaction) MakeTransactionEventReq(
	_eventType tx_lib.TransactionEventEnumType_1,
	_triggerR tx_lib.TriggerReasonEnumType_1,
) (wrappers.CALL, error) {
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
						Value: float64(tx.Evse.EnergyActiveNet_kwh_times100) / 100,
					},
					tx_lib.SampledValueType{
						Measurand: &power_active_import_type,
						UnitOfMeasure: &tx_lib.UnitOfMeasureType{
							Multiplier: 3,
							Unit:       "W",
						},
						Value: float64(tx.Evse.PowerActiveImport_kw_times100) / 100,
					},
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
		SeqNo:     tx.TxSeqNo,
		Timestamp: time.Now().Format(time.RFC3339),
		TransactionInfo: tx_lib.TransactionType{
			TransactionId: tx.Id,
		},
		TriggerReason: _triggerR,
	}

	tx.TxSeqNo += 1

	call_wrapper := wrappers.CALL{
		MessageTypeId: wrappers.CALL_TYPE,
		MessageId:     uuid.New().String(),
		Action:        "TransactionEvent",
		Payload:       *tx_req,
	}

	return call_wrapper, nil

}

// cs.OcppClient.Send(ocppclient.AsyncOcppCall{
// 	Message: *call_wrapper,
// 	SuccessCallback: func(callresult wrappers.CALLRESULT) {
// 		fmt.Println("TransactionEventReq received by CSMS")
// 		onSuccess()
// 	},
// 	ErrorCallback: func(wrappers.CALLERROR) {
// 		log.Error("TransactionEventReq NOT received by CSMS")
// 		onFailure()
// 	},
// })
