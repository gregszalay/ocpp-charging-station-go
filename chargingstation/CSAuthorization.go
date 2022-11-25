package chargingstation

import (
	"bufio"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/gregszalay/ocpp-charging-station-go/ocppclient"
	"github.com/gregszalay/ocpp-messages-go/types/AuthorizeRequest"
	"github.com/gregszalay/ocpp-messages-go/types/AuthorizeResponse"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
	"github.com/sanity-io/litter"
	log "github.com/sirupsen/logrus"
)

func (cs *ChargingStation) authorizeWithRFID(onAuthSuccess func(), onAuthFailure func()) {
	// Create AuthorizeRequest
	authorizeRequest := AuthorizeRequest.AuthorizeRequestJson{
		IdToken: AuthorizeRequest.IdTokenType{
			IdToken: readRFID(),
			Type:    AuthorizeRequest.IdTokenEnumType_1_ISO14443,
		},
	}

	// Wrap AuthorizeRequest in CALL
	call_wrapper := &wrappers.CALL{
		MessageTypeId: wrappers.CALL_TYPE,
		MessageId:     uuid.New().String(),
		Action:        "Authorize",
		Payload:       authorizeRequest,
	}

	// Send AuthorizeRequest and provide the callback implementation
	cs.OcppClient.Send(ocppclient.AsyncOcppCall{
		Message: *call_wrapper,
		SuccessCallback: func(callresult wrappers.CALLRESULT) {
			var resp AuthorizeResponse.AuthorizeResponseJson
			payload_unmarshal_err := resp.UnmarshalJSON(callresult.GetPayloadAsJSON())
			if payload_unmarshal_err != nil {
				fmt.Printf("Failed to unmarshal CALLRESULT message payload. Error: %s", payload_unmarshal_err)
			} else {
				fmt.Println("Payload as an OBJECT:")
				fmt.Println("*******************************")
				litter.Dump(resp)
			}
			if resp.IdTokenInfo.Status != AuthorizeResponse.AuthorizationStatusEnumType_1_Accepted {
				log.Error("Idtoken not accepted, unable to start charging")
				onAuthFailure()
				return
			} else {
				// If the idToken is accepted, start the charging in the EVSE
				onAuthSuccess()
			}
		},
		ErrorCallback: func(wrappers.CALLERROR) {
			log.Error("Failed to send authorize req")
		},
	})
}

func readRFID() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please touch RFID card to reader: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		log.Error("RFID read failed")
		return ""
	}
	fmt.Println("RFID read successfully: ")
	fmt.Println(text)
	return text
}
