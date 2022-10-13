// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gregszalay/ocpp-charging-station-simulator/stationmessages"
	"github.com/gregszalay/ocpp-messages-go/types/StatusNotificationRequest"
	"github.com/gregszalay/ocpp-messages-go/types/TransactionEventRequest"
)

var ocpp_host = flag.String("h", "localhost:3000", "ocpp websocket server host")
var ocpp_url = flag.String("u", "/ocpp", "ocpp URL")
var ocpp_station_id = flag.String("id", "CS001", "id of the charging station")

var simulationinprogress bool = false

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *ocpp_host, Path: *ocpp_url + "/" + *ocpp_station_id}
	fmt.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Printf("Received message: \n%s\n", message)
		}
	}()

	// Sending BootNotificationRequest
	fmt.Println("Creating BootNotificationRequest...")
	boot_req := stationmessages.Create_BootNotificationRequest()
	fmt.Printf("Creating message: \n%s\n", boot_req)
	boot_err := c.WriteMessage(websocket.TextMessage, boot_req)
	if boot_err != nil {
		log.Println("write:", boot_err)
		return
	}

	time.Sleep(time.Second * 5)

	runTXsimulation(c)

	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection.
	close_err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if close_err != nil {
		log.Println("write close:", close_err)
		return
	}

	return

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			fmt.Println("Creating HeartbeatRequest...")
			fmt.Println(t)
			err := c.WriteMessage(websocket.TextMessage, stationmessages.Create_HeartbeatRequest())
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		default:
			if simulationinprogress == false {
				fmt.Print("Please type 'charge' to start charging simulation")
				input := bufio.NewScanner(os.Stdin)
				input.Scan()
				user_input := input.Text()
				fmt.Println(user_input)
				if user_input == "charge" {
					runTXsimulation(c)
				}
			}

		}
	}
}

func runTXsimulation(c *websocket.Conn) {

	simulationinprogress = true
	// Starting transaction - E02 - Cable Plugin First

	// Sending StatusNotificationRequest
	fmt.Println("Creating StatusNotificationRequest...")
	status_req := stationmessages.Create_StatusNotificationRequest(
		StatusNotificationRequest.ConnectorStatusEnumType_1_Occupied)
	fmt.Printf("Sending message: \n%s\n", status_req)
	status_err := c.WriteMessage(websocket.TextMessage, status_req)
	if status_err != nil {
		log.Println("write:", status_err)
		return
	}

	time.Sleep(time.Second * 5)

	// Sending TransactionEventRequest - Start
	fmt.Println("Creating TransactionEventRequest...")
	tx_req := stationmessages.Create_TransactionEventRequest(
		TransactionEventRequest.TransactionEventEnumType_1_Started,
		0,
		22.0,
		TransactionEventRequest.TriggerReasonEnumType_1_CablePluggedIn, "TX004")
	fmt.Printf("Sending message: \n%s\n", tx_req)
	tx_err := c.WriteMessage(websocket.TextMessage, tx_req)
	if tx_err != nil {
		log.Println("write:", tx_err)
		return
	}

	time.Sleep(time.Second * 5)

	// Sending AuthorizeRequest
	fmt.Println("Creating AuthorizeRequest...")
	auth_req := stationmessages.Create_AuthorizeRequest()
	fmt.Printf("Sending message: \n%s\n", auth_req)
	auth_err := c.WriteMessage(websocket.TextMessage, auth_req)
	if auth_err != nil {
		log.Println("write:", auth_err)
		return
	}

	time.Sleep(time.Second * 5)

	// Sending TransactionEventRequest - Update
	fmt.Println("Creating TransactionEventRequest...")
	tx2_req := stationmessages.Create_TransactionEventRequest(
		TransactionEventRequest.TransactionEventEnumType_1_Updated,
		0.1,
		22.0,
		TransactionEventRequest.TriggerReasonEnumType_1_Authorized, "TX004")
	fmt.Printf("Sending message: \n%s\n", tx2_req)
	tx2_err := c.WriteMessage(websocket.TextMessage, tx2_req)
	if tx2_err != nil {
		log.Println("write:", tx2_err)
		return
	}

	time.Sleep(time.Second * 5)

	// Sending TransactionEventRequest - Update
	fmt.Println("Creating TransactionEventRequest...")
	tx3_req := stationmessages.Create_TransactionEventRequest(
		TransactionEventRequest.TransactionEventEnumType_1_Updated,
		0.1,
		22.0,
		TransactionEventRequest.TriggerReasonEnumType_1_ChargingStateChanged, "TX004")
	fmt.Printf("Sending message: \n%s\n", tx3_req)
	tx3_err := c.WriteMessage(websocket.TextMessage, tx3_req)
	if tx3_err != nil {
		log.Println("write:", tx3_err)
		return
	}

	time.Sleep(time.Second * 5)

	// Sending AuthorizeRequest - stop
	fmt.Println("Creating AuthorizeRequest...")
	auth_stop_req := stationmessages.Create_AuthorizeRequest()
	fmt.Printf("Sending message: \n%s\n", auth_stop_req)
	auth_stop_err := c.WriteMessage(websocket.TextMessage, auth_stop_req)
	if auth_err != nil {
		log.Println("write:", auth_stop_err)
		return
	}

	time.Sleep(time.Second * 5)

	// Sending TransactionEventRequest - Update
	fmt.Println("Creating TransactionEventRequest...")
	tx4_req := stationmessages.Create_TransactionEventRequest(
		TransactionEventRequest.TransactionEventEnumType_1_Updated,
		0.1,
		22.0,
		TransactionEventRequest.TriggerReasonEnumType_1_StopAuthorized, "TX004")
	fmt.Printf("Sending message: \n%s\n", tx4_req)
	tx4_err := c.WriteMessage(websocket.TextMessage, tx4_req)
	if tx4_err != nil {
		log.Println("write:", tx4_err)
		return
	}

	time.Sleep(time.Second * 5)

	// Sending StatusNotificationRequest - unplugged
	fmt.Println("Creating StatusNotificationRequest...")
	status_unplugged_req := stationmessages.Create_StatusNotificationRequest(
		StatusNotificationRequest.ConnectorStatusEnumType_1_Available)
	fmt.Printf("Sending message: \n%s\n", status_unplugged_req)
	status_unplugged_err := c.WriteMessage(websocket.TextMessage, status_unplugged_req)
	if status_unplugged_err != nil {
		log.Println("write:", status_unplugged_err)
		return
	}

	time.Sleep(time.Second * 5)

	// Sending TransactionEventRequest - End
	fmt.Println("Creating TransactionEventRequest...")
	tx5_req := stationmessages.Create_TransactionEventRequest(
		TransactionEventRequest.TransactionEventEnumType_1_Ended,
		0.1,
		22.0,
		TransactionEventRequest.TriggerReasonEnumType_1_EVCommunicationLost, "TX004")
	fmt.Printf("Sending message: \n%s\n", tx5_req)
	tx5_err := c.WriteMessage(websocket.TextMessage, tx5_req)
	if tx5_err != nil {
		log.Println("write:", tx5_err)
		return
	}

	time.Sleep(time.Second * 5)

	simulationinprogress = false
}
