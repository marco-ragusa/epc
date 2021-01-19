# External Process Communication

### Description
Encrypted communication between processes through a tcp network stream. The points of contact, 
which are used in the communication between the processes, are the channels of the golang language.

### Status code
- 0 in use
- 1 closed by user (method Close)
- 2 closed by error

## Examples

### Server
 > Before run the server you need to generate the cert and key file using `cert/generate.sh`
```go
package main

import (
	"fmt"
	"log"

	"github.com/marco-ragusa/epc"
)

func main(){
	// func NewStreamServer(port string, bufferSize int, certFile string, keyFile string) *StreamServer
	s := epc.NewStreamServer("8000", 5, "./cert/server.crt", "./cert/server.key")
	// activate the server stream
	if err := s.Start(); err != nil {
		log.Print(err)
	}

	for {
		for conn, data := range s.Connections {
			log.Printf("conn: %s\nmsg: %s\n\n", conn.RemoteAddr(), <-data.Receive)
			data.Send <- fmt.Sprintf("server: data received from conn %s", conn.RemoteAddr())
		}
	}
}
```

### Client
```go
package main

import (
	"log"

	"github.com/marco-ragusa/epc"
)

func main() {
	// func NewStreamClient(host string, port string, bufferSize int) *StreamClient
	c := epc.NewStreamClient("127.0.0.1", "8000", 5)
	// activate the client stream
	if err := c.Start(); err != nil {
		log.Println("Can't connect, conn status: ", c.GetStatus())
		return
	}

	for {
		c.Send <- "test"
		log.Println(<-c.Receive)

		// check if the connection is not open
		if c.GetStatus() > 0 {
			log.Println("Conn status ", c.GetStatus())
			break
		}
	}
}
```
