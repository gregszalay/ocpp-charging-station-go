package stationmessages

import (
	"github.com/google/uuid"
	"github.com/gregszalay/ocpp-messages-go/types/AuthorizeRequest"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
)

func Create_AuthorizeRequest() []byte {
	authorizeRequest := AuthorizeRequest.AuthorizeRequestJson{
		IdToken: AuthorizeRequest.IdTokenType{
			IdToken: "AA00001",
			Type:    AuthorizeRequest.IdTokenEnumType_1_ISO14443,
		},
	}
	call_wrapper := &wrappers.CALL{
		MessageTypeId: wrappers.CALL_TYPE,
		MessageId:     uuid.New().String(),
		Action:        "Authorize",
		Payload:       authorizeRequest,
	}
	json_message := call_wrapper.Marshal()
	return json_message
}
