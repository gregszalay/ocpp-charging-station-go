package evsemanager

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type AsyncEVSEMessage struct {
	Message         string
	SuccessCallback func(string)
}

type EVSE struct {
	Id                               int
	IsEVConnected                    int
	IsChargingEnabled                int
	IsCharging                       int
	IsError                          int
	EnergyActiveNet_wh               int64
	PowerActiveImport_w              int64
	OnEVConnected_fire_once          func()
	OnEVDisconnected_fire_once       func()
	OnEVSEChargingEnabled_fire_once  func()
	OnEVSEChargingDisabled_fire_once func()
	OnEVSEChargingStarted_fire_once  func()
	OnEVSEChargingStopped_fire_once  func()
	OnEVSEError_fire_once            func()
	OnEVSENoError_fire_once          func()
	OnEVConnected_repeat             func()
	OnEVDisconnected_repeat          func()
	OnEVSEChargingEnabled_repeat     func()
	OnEVSEChargingDisabled_repeat    func()
	OnEVSEChargingStarted_repeat     func()
	OnEVSEChargingStopped_repeat     func()
	OnEVSEError_repeat               func()
	OnEVSENoError_repeat             func()
	tcp_conn                         *net.TCPConn
	in_channel                       chan string
	out_channel                      chan string
	lastMessageSentAt                time.Time
	signal                           chan string
}

func CreateAndRunEVSE(id int, servAddr string) (*EVSE, error) {
	// Create new EVSE object
	evse_new := &EVSE{
		// Fill it up with default values
		Id:                               id,
		IsEVConnected:                    0,
		IsChargingEnabled:                0,
		IsCharging:                       0,
		IsError:                          0,
		EnergyActiveNet_wh:               0,
		PowerActiveImport_w:              0,
		OnEVConnected_fire_once:          func() { fmt.Println("EVConnected - No override (fire once)") },
		OnEVDisconnected_fire_once:       func() { fmt.Println("EVDisconnected - No override (fire once)") },
		OnEVSEChargingEnabled_fire_once:  func() { fmt.Println("EVSEChargingEnabled - No override (fire once)") },
		OnEVSEChargingDisabled_fire_once: func() { fmt.Println("EVSEChargingDisabled - No override (fire once)") },
		OnEVSEChargingStarted_fire_once:  func() { fmt.Println("EVSEChargingStarted - No override (fire once)") },
		OnEVSEChargingStopped_fire_once:  func() { fmt.Println("EVSEChargingStopped - No override (fire once)") },
		OnEVConnected_repeat:             func() { fmt.Println("EVConnected - No override") },
		OnEVDisconnected_repeat:          func() { fmt.Println("EVDisconnected - No override") },
		OnEVSEChargingEnabled_repeat:     func() { fmt.Println("EVSEChargingEnabled - No override") },
		OnEVSEChargingDisabled_repeat:    func() { fmt.Println("EVSEChargingDisabled - No override") },
		OnEVSEChargingStarted_repeat:     func() { fmt.Println("EVSEChargingStarted - No override") },
		OnEVSEChargingStopped_repeat:     func() { fmt.Println("EVSEChargingStopped - No override") },
		OnEVSEError_repeat:               func() { fmt.Println("EVSEError - No override") },
		OnEVSENoError_repeat:             func() { fmt.Println("EVSEError - No override") },
		tcp_conn:                         nil,
		in_channel:                       make(chan string, 10),
		out_channel:                      make(chan string, 10),
		lastMessageSentAt:                time.Now(),
		signal:                           make(chan string),
	}

	// Connect to EVSE TCP server
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		return nil, err
	}
	evse_new.tcp_conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("Dial failed:", err.Error())
		return nil, err
	}

	// Authenticate this connection with the EVSE server
	evse_pwd, err := ioutil.ReadFile("evse.pwd")
	if err != nil {
		log.Error("Unable to read EVSE password from file")
		log.Fatal(err)
	}
	_, auth_err := evse_new.tcp_conn.Write([]byte(string(evse_pwd) + "\n"))
	if auth_err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}

	// LISTEN
	go func() { // listen for incoming messages and put them into a queue
		reply := make([]byte, 50)
		for {
			n, err := evse_new.tcp_conn.Read(reply)
			if err != nil {
				log.Error("TCP read failed:", err.Error())
				os.Exit(1)
			}
			if n != 0 {
				reply_str_raw := string(reply[:n])
				reply_str := strings.TrimSpace(reply_str_raw)
				if reply_str != "" {
					log.Info("reply from server = ", reply_str)
					evse_new.in_channel <- reply_str
				}
				reply = make([]byte, 50)
			}
			//reply = nil
			//reply = make([]byte, 50)
		}
	}()

	// PROCESS
	go func() { // keep looking through incoming message queue and process the messages
		for evse_reply := range evse_new.in_channel {
			evse_new.processEVSEMessage(evse_reply)
		}
		log.Info("PROCESS goroutine has finished")
	}()

	// SEND
	go func() { // keep looking for messages to send, send message
		ticker_status := time.NewTicker(time.Millisecond * 250)
		defer ticker_status.Stop()
		for {
			select {
			case t := <-ticker_status.C:
				_ = t
				if len(evse_new.out_channel) == 0 {
					break
				}
				log.Debug("Current time: ", time.Now())
				new_mess := <-evse_new.out_channel
				log.Debug("writing the following message to EVSE controller: ", new_mess)
				_, err := evse_new.tcp_conn.Write([]byte(new_mess))
				if err != nil {
					println("Write to server failed:", err.Error())
					os.Exit(1)
				}
			default:
			}
		}
	}()

	// EVSE POLLING
	ticker_status := time.NewTicker(time.Millisecond * 300)
	go func() {
		defer ticker_status.Stop()
		for {
			select {
			case t := <-ticker_status.C:
				_ = t
				if len(evse_new.out_channel) == 0 { // TODO implement proper limit
					evse_new.out_channel <- "status?\n"
					evse_new.out_channel <- "metervalues?\n"
				}
			default:
			}
		}
	}()

	return evse_new, nil
}

func (evse *EVSE) EnableCharging() {
	evse.out_channel <- "start\n"
}

func (evse *EVSE) DisableCharging() {
	evse.out_channel <- "stop\n"
}

func (evse *EVSE) processEVSEMessage(evse_reply string) {
	log.Trace("Processing string from EVSE:", evse_reply)
	split_result := strings.Split(evse_reply, ":")
	if len(split_result) < 2 {
		log.Debug("Unable to update status, status string is length is less than 2")
		return
	}
	message_header := strings.Trim(split_result[0], " ")
	message_body := strings.Trim(split_result[1], " ")

	log.Trace("message_header: ", message_header)
	log.Trace("message_body", message_body)

	switch message_header {
	case "status":
		evse.updateStatus(message_body)
	case "metervalues":
		evse.updateMeterValues(message_body)
	default:
		log.Warning("Received unknown message type from EVSE")
	}
}

func (evse *EVSE) Disconnect() {
	evse.tcp_conn.Close()
}

func (evse *EVSE) updateStatus(statusString string) {
	log.Trace("Original status string: ", statusString)
	split_result := strings.Split(statusString, ",")
	if len(split_result) < 4 {
		log.Debug("Unable to update status, status string is less than 4")
		return
	}

	IsEVConnected_new_str := strings.Trim(split_result[0], " ,")
	IsChargingEnabled_new_str := strings.Trim(split_result[1], " ,")
	IsCharging_new_str := strings.Trim(split_result[2], " ,")
	IsError_new_str := strings.Trim(split_result[3], " ,\n")

	log.Trace("split_result[0]", IsEVConnected_new_str)
	log.Trace("split_result[1]", IsChargingEnabled_new_str)
	log.Trace("split_result[2]", IsCharging_new_str)
	log.Trace("split_result[3]", IsError_new_str)

	if IsEVConnected, err := strconv.ParseInt(IsEVConnected_new_str, 10, 64); err != nil {
		log.Error("unable to convert IsEVConnected to int", err)
	} else {
		if IsEVConnected == 1 && evse.IsEVConnected == 0 {
			evse.IsEVConnected = 1
			evse.OnEVConnected_repeat()
			evse.OnEVConnected_fire_once()
			evse.OnEVConnected_fire_once = func() {}
		}
		if IsEVConnected == 0 && evse.IsEVConnected == 1 {
			evse.IsEVConnected = 0
			evse.OnEVDisconnected_repeat()
			evse.OnEVDisconnected_fire_once()
			evse.OnEVDisconnected_fire_once = func() {}
		}
	}

	if IsChargingEnabled, err := strconv.ParseInt(IsChargingEnabled_new_str, 10, 64); err != nil {
		log.Error("unable to convert IsChargingEnabled to int", err)
	} else {
		if IsChargingEnabled == 1 && evse.IsChargingEnabled == 0 {
			evse.IsChargingEnabled = 1
			evse.OnEVSEChargingEnabled_repeat()
			evse.OnEVSEChargingEnabled_fire_once()
			evse.OnEVSEChargingEnabled_fire_once = func() {}
		}
		if IsChargingEnabled == 0 && evse.IsChargingEnabled == 1 {
			evse.IsChargingEnabled = 0
			evse.OnEVSEChargingDisabled_repeat()
			evse.OnEVSEChargingDisabled_fire_once()
			evse.OnEVSEChargingDisabled_fire_once = func() {}
		}
	}

	if IsCharging, err := strconv.ParseInt(IsCharging_new_str, 10, 64); err != nil {
		log.Error("unable to convert IsCharging to int", err)
	} else {
		if IsCharging == 1 && evse.IsCharging == 0 {
			evse.IsCharging = 1
			evse.OnEVSEChargingStarted_repeat()
			evse.OnEVSEChargingStarted_fire_once()
			evse.OnEVSEChargingStarted_fire_once = func() {}

		}
		if IsCharging == 0 && evse.IsCharging == 1 {
			evse.IsCharging = 0
			evse.OnEVSEChargingStopped_repeat()
			evse.OnEVSEChargingStopped_fire_once()
			evse.OnEVSEChargingStopped_fire_once = func() {}
		}
	}

	if IsError, err := strconv.ParseInt(IsError_new_str, 10, 64); err != nil {
		log.Error("unable to convert IsError to int", err)
	} else {
		if IsError == 1 && evse.IsError == 0 {
			evse.IsError = 1
			evse.OnEVSEError_repeat()
			evse.OnEVSEError_fire_once()
			evse.OnEVSEError_fire_once = func() {}
		}
		if IsError == 0 && evse.IsError == 1 {
			evse.IsError = 0
			evse.OnEVSENoError_repeat()
			evse.OnEVSENoError_fire_once()
			evse.OnEVSENoError_fire_once = func() {}
		}
	}

}

func (evse *EVSE) updateMeterValues(meterValuesString string) {
	log.Trace("Original meterValues string: ", meterValuesString)
	split_result := strings.Split(meterValuesString, ",")
	if len(split_result) < 2 {
		log.Error("Unable to update metervalues, meterValuesString is less than 2")
		return
	}

	EnergyActiveNet_wh_str := strings.Trim(split_result[0], " ,")
	PowerActiveImport_w_str := strings.Trim(split_result[1], " ,")

	if EnergyActiveNet_wh, err := strconv.ParseInt(EnergyActiveNet_wh_str, 10, 64); err != nil {
		log.Error("unable to convert EnergyActiveNet_wh to int", err)
	} else {
		evse.EnergyActiveNet_wh = EnergyActiveNet_wh
	}

	if PowerActiveImport_w, err := strconv.ParseInt(PowerActiveImport_w_str, 10, 64); err != nil {
		log.Error("unable to convert PowerActiveImport_w to int", err)
	} else {
		evse.PowerActiveImport_w = PowerActiveImport_w
	}

}
