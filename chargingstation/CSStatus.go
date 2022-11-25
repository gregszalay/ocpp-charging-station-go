package chargingstation

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gregszalay/ocpp-charging-station-go/evsemanager"
	"github.com/gregszalay/ocpp-charging-station-go/ocppclient"
	"github.com/gregszalay/ocpp-messages-go/types/StatusNotificationRequest"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
	log "github.com/sirupsen/logrus"
)

func (cs *ChargingStation) SendStatusNotification(evse evsemanager.EVSE) {
	
	// Fetch status of the EVSE
	var status =  StatusNotificationRequest.ConnectorStatusEnumType_1_Available
	if evse.IsError == 1 {
		status = StatusNotificationRequest.ConnectorStatusEnumType_1_Faulted
	}
	if evse.IsEVConnected == 1 {
		status = StatusNotificationRequest.ConnectorStatusEnumType_1_Occupied
	}

	// Create StatusNotificationRequest
	statusNotificationRequest := &StatusNotificationRequest.StatusNotificationRequestJson{
		ConnectorId:     0, //TODO check numbering standard
		ConnectorStatus: status,
		EvseId:          evse.Id,
		Timestamp:       time.Now().Format(time.RFC3339),
	}
	call_wrapper := &wrappers.CALL{
		MessageTypeId: wrappers.CALL_TYPE,
		MessageId:     uuid.New().String(),
		Action:        "StatusNotification",
		Payload:       *statusNotificationRequest,
	}

	// Send AuthorizeRequest and provide the callback implementation
	cs.OcppClient.Send(ocppclient.AsyncOcppCall{
		Message: *call_wrapper,
		SuccessCallback: func(callresult wrappers.CALLRESULT) {
			fmt.Println("Statusnotification received by CSMS")
		},
		ErrorCallback: func(wrappers.CALLERROR) {
			log.Error("Statusnotification NOT received by CSMS")
		},
	})
}
