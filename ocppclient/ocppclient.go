package ocppclient

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
	"github.com/sanity-io/litter"
	log "github.com/sirupsen/logrus"
)

type AsyncOcppCall struct {
	Message         wrappers.CALL
	SuccessCallback func(wrappers.CALLRESULT)
	ErrorCallback   func(wrappers.CALLERROR)
}

var calls_to_send chan AsyncOcppCall
var calls_awaiting_response map[string]AsyncOcppCall
var incoming_call_handlers map[string]func(wrappers.CALL)

func Connect(u url.URL, handlers map[string]func(wrappers.CALL)) {

	// Initialize the callresult handlers
	incoming_call_handlers = handlers

	// Create the WS connection
	ws_conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer ws_conn.Close()

	// Start the threads
	go runInbox(ws_conn)
	go runOutbox(ws_conn)
}

func Disconnect() {

}

func Send(call AsyncOcppCall) {
	calls_to_send <- call
}

func runOutbox(ws_conn *websocket.Conn) {
	// initialize the outbound message channel
	calls_to_send = make(chan AsyncOcppCall)
	for ocpp_call := range calls_to_send {
		err := ws_conn.WriteMessage(websocket.TextMessage, ocpp_call.Message.Marshal())
		if err != nil {
			log.Println("write:", err)
			return
		}
		calls_awaiting_response[ocpp_call.Message.MessageId] = ocpp_call
	}

}

func runInbox(ws_conn *websocket.Conn) {
	calls_awaiting_response = make(map[string]AsyncOcppCall)

	for {
		_, message, err := ws_conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		fmt.Printf("\nReceived message: \n%s\n", message)
		processIncomingMessage(message)
	}
}

func processIncomingMessage(message []byte) {
	messageTypeId, action, err := parseMessageTypeIdAction(message)
	if err != nil {
		log.Error("could not parse message type id")
	}

	switch messageTypeId {
	case wrappers.CALL_TYPE:
		var call wrappers.CALL
		call_unmarshal_err := call.UnmarshalJSON(message)
		if call_unmarshal_err != nil {
			fmt.Printf("Failed to unmarshal OCPP CALL message. Error: %s", call_unmarshal_err)
		} else {
			fmt.Println("OCPP CALL message as an OBJECT:")
			fmt.Println("*******************************")
			litter.Dump(call)
		}
		// invoke callback
		(incoming_call_handlers[action])(call)
	case wrappers.CALLRESULT_TYPE:
		var callresult wrappers.CALLRESULT
		call_result_unmarshal_err := callresult.UnmarshalJSON(message)
		if call_result_unmarshal_err != nil {
			fmt.Printf("Failed to unmarshal OCPP CALLRESULT message. Error: %s", call_result_unmarshal_err)
		} else {
			fmt.Println("OCPP CALLRESULT message as an OBJECT:")
			fmt.Println("*******************************")
			litter.Dump(callresult)
		}
		// invoke callback
		calls_awaiting_response[callresult.MessageId].SuccessCallback(callresult)
	case wrappers.CALLERROR_TYPE:
		var callerror wrappers.CALLERROR
		callerror_result_unmarshal_err := callerror.UnmarshalJSON(message)
		if callerror_result_unmarshal_err != nil {
			fmt.Printf("Failed to unmarshal OCPP CALLERROR message. Error: %s", callerror_result_unmarshal_err)
		} else {
			fmt.Println("OCPP CALLERROR message as an OBJECT:")
			fmt.Println("*******************************")
			litter.Dump(callerror)
		}
		// invoke callback
		calls_awaiting_response[callerror.MessageId].ErrorCallback(callerror)
	}
}

func parseMessageTypeIdAction(message []byte) (int, string, error) {
	var data []interface{}
	err := json.Unmarshal([]byte(message), &data)
	if err != nil {
		log.Error("could not unmarshal json", err)
		return 0, "", err
	}
	messageTypeId, ok := data[0].(float64)
	if !ok {
		log.Error("data[0] is not a uint8", err)
		return 0, "", err
	}
	action, err_action := data[2].(string)
	if !err_action || action == "" {
		return log.Error("CALL data[2] is not a string", err_action)
	}

	return int(messageTypeId), action, nil
}
