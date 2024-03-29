package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	UNIX_SOCK_PIPE_PATH = "/var/run/unixsock_test.sock" // socket file path
)

var (
	exitSemaphore chan bool
)

func main() {
	// Get unix socket address based on file path
	uaddr, err := net.ResolveUnixAddr("unix", UNIX_SOCK_PIPE_PATH)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Connect server with unix socket
	uconn, err := net.DialUnix("unix", nil, uaddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Close unix socket when exit this function
	defer uconn.Close()

	// Wait to receive response
	go onMessageReceived(uconn)

	// Send a request to server
	// you can define your own rules
	msg := "msgtell me current time\n"
	sendRequest(uconn, []byte(msg))

	//msg Wait server response
	// change this duration bigger than server sleep time to get correct response
	exitSemaphore = make(chan bool)
	select {
	case <-time.After(time.Duration(2) * time.Second):
		fmt.Println("Wait response timeout")
	case <-exitSemaphore:
		fmt.Println("Get response correctly")
	}

	close(exitSemaphore)
}

/*******************************************************
 * Send request to server, you can define your own proxy
 * conn: conn handler
 *******************************************************/
func sendRequest(conn *net.UnixConn, data []byte) {
	buf := new(bytes.Buffer)
	msglen := uint32(len(data))

	binary.Write(buf, binary.BigEndian, &msglen)
	data = append(buf.Bytes(), data...)

	conn.Write(data)
}

/*******************************************************
 * Handle connection and response
 * conn: conn handler
 *******************************************************/
func onMessageReceived(conn *net.UnixConn) {
	//for { // io Read will wait here, we don't need for loop to check
	//      Read information from response
	data, err := parseResponse(conn)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%v\tReceived from server: %s\n", time.Now(), string(data))
	}

	// Exit when receiveive data from server
	exitSemaphore <- true
	//}
}

/*****************exitSemaphore**************************************
* Parse request of unix socket
* conn: conn handler
*******************************************************/
func parseResponse(conn *net.UnixConn) ([]byte, error) {
	var reqLen uint32
	lenBytes := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBytes); err != nil {
		return nil, err
	}

	lenBuf := bytes.NewBuffer(lenBytes)
	if err := binary.Read(lenBuf, binary.BigEndian, &reqLen); err != nil {
		return nil, err
	}

	reqBytes := make([]byte, reqLen)
	_, err := io.ReadFull(conn, reqBytes)

	if err != nil {
		return nil, err
	}

	return reqBytes, nil
}
