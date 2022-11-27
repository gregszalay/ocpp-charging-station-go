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

func CreateAndRunOCPPClient(_csms_url url.URL) (*OCPPClient, error) {
	// Create new OCPPClient
	ocpp_client_new := &OCPPClient{
		calls_to_send:           make(chan AsyncOcppCall, 10),   // Initialize the outbound message channel
		calls_awaiting_response: make(map[string]AsyncOcppCall), // Initialize sent call message storage
		Calls_received:          make(chan wrappers.CALL, 10),
		ws_conn:                 nil,
	}

	ws_conn_new, _, err := websocket.DefaultDialer.Dial(_csms_url.String(), nil)
	if err != nil {
		log.Info("2 1 1")
		log.Fatal("dial:", err)
		return nil, err
	}

	ocpp_client_new.ws_conn = ws_conn_new

	// LISTEN
	go func() { // listen for incoming messages and put them into a queue
		for {
			_, message, err := ocpp_client_new.ws_conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Printf("\nReceived message: \n%s\n", message)
			ocpp_client_new.processIncomingMessage(message)
		}
		log.Info("LISTEN goroutine has finished")
	}()

	// PROCESS
	// go func() { // keep looking through incoming message queue and process the messages
	// 	for incoming_message := range ocpp_client_new.Calls_received {
	// 		ocpp_client_new.processIncomingMessage(incoming_message)
	// 	}
	// 	log.Info("PROCESS goroutine has finished")
	// }()

	// SEND
	go func() { // keep looking for messages to send, send message
		ticker_status := time.NewTicker(time.Second * 1)
		defer ticker_status.Stop()
		for {
			select {
			case t := <-ticker_status.C:
				_ = t
				if len(ocpp_client_new.calls_to_send) == 0 {
					break
				}
				fmt.Println("Current time: ", time.Now())
				new_message := <-ocpp_client_new.calls_to_send
				err := ocpp_client_new.ws_conn.WriteMessage(websocket.TextMessage, new_message.Message.Marshal())
				if err != nil {
					log.Println("write:", err)
					return
				}
				new_message.SentAt = time.Now()
				ocpp_client_new.calls_awaiting_response[new_message.Message.MessageId] = new_message
			default:
			}
		}
		log.Info("SEND goroutine has finished")
	}()

	go func() {
		for {
			for messageId, sent_ocpp_call := range ocpp_client_new.calls_awaiting_response {
				if time.Since(sent_ocpp_call.SentAt) > time.Millisecond*3000 {
					delete(ocpp_client_new.calls_awaiting_response, messageId) // delete the timed out message
				}
			}
			//time.Sleep(time.Millisecond * 500)
		}
	}()

	return ocpp_client_new, nil
}

func (cl *OCPPClient) Disconnect() {
	cl.ws_conn.Close()
}

func (cl *OCPPClient) Send(call AsyncOcppCall) {
	cl.calls_to_send <- call
}

func (cl *OCPPClient) processIncomingMessage(message []byte) {
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
		cl.Calls_received <- call
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
		cl.calls_awaiting_response[callresult.MessageId].SuccessCallback(callresult)
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
		cl.calls_awaiting_response[callerror.MessageId].ErrorCallback(callerror)
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
