package epc

import (
	"crypto/tls"
	"encoding/gob"
	"net"
)

func startListener(ss *StreamServer, lsn net.Listener) {
	// close the listener at the finish
	defer func() { _ = lsn.Close() }()

	// accept connections
	for {
		conn, err := lsn.Accept()
		if err != nil {
			ss.status = 2
			return
		}

		ss.Connections[conn] = Data{
			make(chan string, ss.bufferSize),
			make(chan string, ss.bufferSize),
		}

		go startServerReceiver(ss, conn)
		go startServerSender(ss, conn)
	}
}

func startServerReceiver(ss *StreamServer, conn net.Conn) {
	// close the connection at the finish
	defer func() {
		// remove conn from the map of Connections
		delete(ss.Connections, conn)
		_ = conn.Close()
	}()

	// structure of the message
	var s Stream
	// initialize the gob decoder
	dec := gob.NewDecoder(conn)

	for {
		if err := dec.Decode(&s); err != nil {
			// receiving empty strings, this avoids blocking the program
			ss.Connections[conn].Receive <- ""
			return
		}

		ss.Connections[conn].Receive <- s.Msg

		// close server connection if status is greater of 0
		if ss.status > 0 {
			return
		}
	}
}

func startServerSender(ss *StreamServer, conn net.Conn) {
	// close the connection at the finish
	defer func() { _ = conn.Close() }()

	// structure of the message
	var s Stream
	// initialize the gon encoder
	encoder := gob.NewEncoder(conn)

	for {
		// close connection to the server
		if ss.status > 0 {
			return
		}

		// get data to send
		s.Msg = <-ss.Connections[conn].Send

		// send stream through the network
		if err := encoder.Encode(&s); err != nil {
			return
		}
	}
}

// Data struct of send and receive channel
type Data struct {
	Send    chan string
	Receive chan string
}

// StreamServer structure
type StreamServer struct {
	port       string
	status     int
	certFile   string
	keyFile    string
	bufferSize int

	Connections map[net.Conn]Data
}

// NewStreamServer constructor
func NewStreamServer(port string, bufferSize int, certFile string, keyFile string) *StreamServer {
	return &StreamServer{
		port:       port,
		bufferSize: bufferSize,
		certFile:   certFile,
		keyFile:    keyFile,

		Connections: make(map[net.Conn]Data),
	}
}

// Start connection
func (ss *StreamServer) Start() error {
	// reset status if you run it another time
	ss.status = 0
	cer, err := tls.LoadX509KeyPair(ss.certFile, ss.keyFile)
	if err != nil {
		ss.status = 2
		return err
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	// initialize and start the listener
	lsn, err := tls.Listen("tcp", ":"+ss.port, config)
	if err != nil {
		ss.status = 2
		return err
	}
	go startListener(ss, lsn)

	return nil
}

// Close connection
func (ss *StreamServer) Close() {
	ss.status = 1
}

// GetStatus of the connection
// 2 server closed by error
// 1 server closed by user, Close method
// 0 server in use
func (ss *StreamServer) GetStatus() int {
	return ss.status
}
