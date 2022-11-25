package ocppclient

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
	"github.com/sanity-io/litter"
	log "github.com/sirupsen/logrus"
)

type AsyncOcppCall struct {
	Message         wrappers.CALL
	SuccessCallback func(wrappers.CALLRESULT)
	ErrorCallback   func(wrappers.CALLERROR)
	SentAt          time.Time
}

type OCPPClient struct {
	calls_to_send           chan AsyncOcppCall
	calls_awaiting_response map[string]AsyncOcppCall
	Calls_received          chan wrappers.CALL
	ws_conn                 *websocket.Conn
}

func CreateOCPPClient() (*OCPPClient, error) {
	// Create new OCPPClient
	new_ocpp_client := &OCPPClient{
		calls_to_send:           make(chan AsyncOcppCall),       // Initialize the outbound message channel
		calls_awaiting_response: make(map[string]AsyncOcppCall), // Initialize sent call message storage
		Calls_received:          make(chan wrappers.CALL),
		ws_conn:                 nil,
	}
	return new_ocpp_client, nil
}

func (client *OCPPClient) ConnectToCSMS(_csms_url url.URL) {
	// Create the WS connection
	ws_conn_new, _, err := websocket.DefaultDialer.Dial(_csms_url.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	client.ws_conn = ws_conn_new
	defer client.ws_conn.Close()
	// Start the threads
	go client.runInbox()
	go client.runOutbox()
	go client.timeoutCheckLoop()
}

func (client *OCPPClient) Disconnect() {
	client.ws_conn.Close()
}

func (client *OCPPClient) Send(call AsyncOcppCall) {
	client.calls_to_send <- call
}

func (client *OCPPClient) runOutbox() {
	for ocpp_call := range client.calls_to_send {
		err := client.ws_conn.WriteMessage(websocket.TextMessage, ocpp_call.Message.Marshal())
		if err != nil {
			log.Println("write:", err)
			return
		}
		ocpp_call.SentAt = time.Now()
		client.calls_awaiting_response[ocpp_call.Message.MessageId] = ocpp_call
	}
}

func (client *OCPPClient) timeoutCheckLoop() {
	for {
		for messageId, sent_ocpp_call := range client.calls_awaiting_response {
			if time.Since(sent_ocpp_call.SentAt) > time.Millisecond*3000 {
				delete(client.calls_awaiting_response, messageId) // delete the timed out message
			}
		}
		time.Sleep(time.Millisecond * 500)
	}
}

func (client *OCPPClient) runInbox() {
	for {
		_, message, err := client.ws_conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		fmt.Printf("\nReceived message: \n%s\n", message)
		client.processIncomingMessage(message)
	}
}

func (client *OCPPClient) processIncomingMessage(message []byte) {
	messageTypeId, err := parseMessageTypeId(message)
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
		client.Calls_received <- call
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
		client.calls_awaiting_response[callresult.MessageId].SuccessCallback(callresult)
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
		client.calls_awaiting_response[callerror.MessageId].ErrorCallback(callerror)
	}
}

func parseMessageTypeId(message []byte) (int, error) {
	var data []interface{}
	err := json.Unmarshal([]byte(message), &data)
	if err != nil {
		log.Error("could not unmarshal json", err)
		return 0, err
	}
	messageTypeId, ok := data[0].(float64)
	if !ok {
		log.Error("data[0] is not a uint8", err)
		return 0, err
	}

	return int(messageTypeId), nil
}
