package stationmessages

import (
	"github.com/google/uuid"
	"github.com/gregszalay/ocpp-messages-go/types/HeartbeatRequest"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
)

func Create_HeartbeatRequest() []byte {
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
	json_message := call_wrapper.Marshal()
	return json_message
}
