package evse

import (
	"net"
	"os"
)

type AsyncEVSEMessage struct {
	Message         string
	SuccessCallback func(string)
}

var message_awaiting_response AsyncEVSEMessage
var conn *net.TCPConn

func Connect(servAddr string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}

}

func Disconnect() {
	conn.Close()
}

func RunInbox() {
	reply := make([]byte, 1024)
	for {
		_, err := conn.Read(reply)
		if err != nil {
			println("TCP read failed:", err.Error())
			os.Exit(1)
		}
		println("reply from server=", string(reply))
		// invoke callback
		message_awaiting_response.SuccessCallback(string(reply))
		// empty the message holder
		message_awaiting_response = AsyncEVSEMessage{}
	}
}

func Send(message string) {
	println("writing the following message to EVSE controller: ", message)
	_, err := conn.Write([]byte(message))
	if err != nil {
		println("Write to server failed:", err.Error())
		os.Exit(1)
	}
}
