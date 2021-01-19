package epc

import (
	"crypto/tls"
	"encoding/gob"
	"net"
)

func startClientSender(sc *StreamClient, conn net.Conn) {
	// close the connection at the finish
	defer func() { _ = conn.Close() }()

	// structure of the message
	var s Stream
	// initialize the gon encoder
	encoder := gob.NewEncoder(conn)

	for {
		// close connection to the server
		if sc.status > 0 {
			return
		}

		// get data to send
		s.Msg = <-sc.Send

		// send stream through the network
		if err := encoder.Encode(&s); err != nil {
			sc.status = 2
			return
		}
	}
}

func startClientReceiver(sc *StreamClient, conn net.Conn) {
	// close the connection at the finish
	defer func() { _ = conn.Close() }()

	// structure of the message
	var s Stream
	// initialize the gob decoder
	dec := gob.NewDecoder(conn)

	for {
		if err := dec.Decode(&s); err != nil {
			// set status err
			sc.status = 2
			// receiving empty strings, this avoids blocking the program
			sc.Receive <- ""
			return
		}

		sc.Receive <- s.Msg

		// close server connection if status is greater of 0
		if sc.status > 0 {
			return
		}
	}
}

// StreamClient structure
type StreamClient struct {
	host   string
	port   string
	status int

	Send    chan string
	Receive chan string
}

// NewStreamClient constructor
func NewStreamClient(host string, port string, bufferSize int) *StreamClient {
	return &StreamClient{
		host:    host,
		port:    port,
		Send:    make(chan string, bufferSize),
		Receive: make(chan string, bufferSize),
	}
}

// Start connection
func (sc *StreamClient) Start() error {
	// reset status if you run it another time
	sc.status = 0
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	// start the connection to the server
	conn, err := tls.Dial("tcp", sc.host+":"+sc.port, conf)
	if err != nil {
		sc.status = 2
		return err
	}
	go startClientSender(sc, conn)
	go startClientReceiver(sc, conn)

	return nil
}

// Close connection
func (sc *StreamClient) Close() {
	sc.status = 1
}

// GetStatus of the connection
// 2 connection closed by error
// 1 connection closed by user, Close method
// 0 connection in use
func (sc *StreamClient) GetStatus() int {
	return sc.status
}
