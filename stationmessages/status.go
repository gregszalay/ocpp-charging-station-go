package stationmessages

import (
	"time"

	"github.com/google/uuid"
	"github.com/gregszalay/ocpp-messages-go/types/StatusNotificationRequest"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
)

func Create_StatusNotificationRequest(status StatusNotificationRequest.ConnectorStatusEnumType_1) []byte {
	statusNotificationRequest := &StatusNotificationRequest.StatusNotificationRequestJson{
		ConnectorId:     1,
		ConnectorStatus: status,
		EvseId:          1,
		Timestamp:       time.Now().Format(time.RFC3339),
	}
	call_wrapper := &wrappers.CALL{
		MessageTypeId: wrappers.CALL_TYPE,
		MessageId:     uuid.New().String(),
		Action:        "StatusNotification",
		Payload:       *statusNotificationRequest,
	}
	json_message := call_wrapper.Marshal()
	return json_message
}
