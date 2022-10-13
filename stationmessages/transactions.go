package stationmessages

import (
	"time"

	"github.com/google/uuid"
	tx "github.com/gregszalay/ocpp-messages-go/types/TransactionEventRequest"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
)

var seq_number int = 1

func Create_TransactionEventRequest(
	_eventType tx.TransactionEventEnumType_1,
	_energy float64,
	_power float64,
	_triggerR tx.TriggerReasonEnumType_1,
	_transactionId string) []byte {
	energy_active_net_type := tx.MeasurandEnumType_1_EnergyActiveNet
	power_active_import_type := tx.MeasurandEnumType_1_PowerActiveImport
	tx_req := &tx.TransactionEventRequestJson{
		EventType: _eventType,
		MeterValue: []tx.MeterValueType{
			tx.MeterValueType{
				SampledValue: []tx.SampledValueType{
					tx.SampledValueType{
						Measurand: &energy_active_net_type,
						UnitOfMeasure: &tx.UnitOfMeasureType{
							Multiplier: 0,
							Unit:       "Wh",
						},
						Value: _energy,
					},
					tx.SampledValueType{
						Measurand: &power_active_import_type,
						UnitOfMeasure: &tx.UnitOfMeasureType{
							Multiplier: 3,
							Unit:       "W",
						},
						Value: _power,
					},
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
		SeqNo:     seq_number,
		Timestamp: time.Now().Format(time.RFC3339),
		TransactionInfo: tx.TransactionType{
			TransactionId: _transactionId,
		},
		TriggerReason: _triggerR,
	}

	seq_number += 1

	call_wrapper := &wrappers.CALL{
		MessageTypeId: wrappers.CALL_TYPE,
		MessageId:     uuid.New().String(),
		Action:        "TransactionEvent",
		Payload:       *tx_req,
	}

	json_message := call_wrapper.Marshal()
	return json_message
}
