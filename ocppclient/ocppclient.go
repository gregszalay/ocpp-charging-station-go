package ocppclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gregszalay/ocpp-messages-go/wrappers"
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
	mu                      sync.Mutex
}

func CreateAndRunOCPPClient(_csms_url url.URL) (*OCPPClient, error) {
	// Create new OCPPClient
	ocpp_client_new := &OCPPClient{
		calls_to_send:           make(chan AsyncOcppCall, 100),  // Initialize the outbound message channel
		calls_awaiting_response: make(map[string]AsyncOcppCall), // Initialize sent call message storage
		Calls_received:          make(chan wrappers.CALL, 100),
		ws_conn:                 nil,
	}

	// Read the key pair to create certificate
	cert, err := tls.LoadX509KeyPair("client_cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}

	// Create a CA certificate pool and add cert.pem to it
	caCert, err := ioutil.ReadFile("client_cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			ClientCAs:    caCertPool,
			Certificates: []tls.Certificate{cert},
		},
	}

	ws_conn_new, _, err := dialer.Dial(_csms_url.String(), nil)
	if err != nil {
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
			ocpp_client_new.mu.Lock()
			ocpp_client_new.processIncomingMessage(message)
			ocpp_client_new.mu.Unlock()
		}
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
		for {
			for message := range ocpp_client_new.calls_to_send {
				fmt.Println("Current time: ", time.Now())
				log.Info("==> Sending CALL message to CSMS")
				log.Info(string(message.Message.Marshal()))
				err := ocpp_client_new.ws_conn.WriteMessage(websocket.TextMessage, message.Message.Marshal())
				if err != nil {
					log.Println("write:", err)
					return
				}
				message.SentAt = time.Now()
				ocpp_client_new.mu.Lock()
				ocpp_client_new.calls_awaiting_response[message.Message.MessageId] = message
				ocpp_client_new.mu.Unlock()
				//time.Sleep(time.Millisecond * 10)
			}
		}
	}()

	go func() {
		for {
			ocpp_client_new.mu.Lock()
			for messageId, sent_ocpp_call := range ocpp_client_new.calls_awaiting_response {
				if time.Since(sent_ocpp_call.SentAt) > time.Millisecond*10000 {
					delete(ocpp_client_new.calls_awaiting_response, messageId) // delete the timed out message
				}
			}
			ocpp_client_new.mu.Unlock()
			//time.Sleep(time.Millisecond * 50)
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
		return
	}

	switch messageTypeId {
	case wrappers.CALL_TYPE:
		var call wrappers.CALL
		call_unmarshal_err := call.UnmarshalJSON(message)
		if call_unmarshal_err != nil {
			fmt.Printf("Failed to unmarshal OCPP CALL message. Error: %s", call_unmarshal_err)
		} else {
			// fmt.Println("OCPP CALL message as an OBJECT:")
			// fmt.Println("*******************************")
			// litter.Dump(call)
		}
		cl.Calls_received <- call
		log.Info("<== Received CALL message from CSMS")
		log.Info(string(call.Marshal()))
	case wrappers.CALLRESULT_TYPE:
		var callresult wrappers.CALLRESULT
		call_result_unmarshal_err := callresult.UnmarshalJSON(message)
		if call_result_unmarshal_err != nil {
			fmt.Printf("Failed to unmarshal OCPP CALLRESULT message. Error: %s", call_result_unmarshal_err)
		} else {
			// fmt.Println("OCPP CALLRESULT message as an OBJECT:")
			// fmt.Println("*******************************")
			// litter.Dump(callresult)
		}
		// invoke callback
		if val, ok := cl.calls_awaiting_response[callresult.MessageId]; ok {
			if val.SuccessCallback == nil {
				log.Error("callresult successcallback is nil")
				return
			}
			//do something here
			val.SuccessCallback(callresult)
			log.Info("<== Received CALLRESULT message from CSMS")
			log.Info(string(callresult.Marshal()))
		} else {
			log.Error("callresult successcallback does not exist")
			return
		}
	case wrappers.CALLERROR_TYPE:
		var callerror wrappers.CALLERROR
		callerror_result_unmarshal_err := callerror.UnmarshalJSON(message)
		if callerror_result_unmarshal_err != nil {
			fmt.Printf("Failed to unmarshal OCPP CALLERROR message. Error: %s", callerror_result_unmarshal_err)
		} else {
			// fmt.Println("OCPP CALLERROR message as an OBJECT:")
			// fmt.Println("*******************************")
			// litter.Dump(callerror)
		}
		// invoke callback
		if val, ok := cl.calls_awaiting_response[callerror.MessageId]; ok {
			//do something here
			if val.ErrorCallback == nil {
				log.Error("callerror errorcallback is nil")
				return
			}
			val.ErrorCallback(callerror)
			log.Info("<== Received CALLERROR message from CSMS")
			log.Info(string(callerror.Marshal()))
		} else {

			log.Error("callerror errorcallback does not exist")
		}
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
