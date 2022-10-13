package stationmessages

import (
	"github.com/google/uuid"
	"github.com/gregszalay/ocpp-messages-go/types/BootNotificationRequest"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
)

func Create_BootNotificationRequest() []byte {
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
	json_message := call_wrapper.Marshal()
	return json_message
}
