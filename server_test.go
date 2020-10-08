package net

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTCPServer(t *testing.T) {
	tcpServer := NewTCPServer()

	tcpServer.SetAddress(":9999")
	err := tcpServer.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	time.Sleep(1 * time.Second)

	tcpClient := NewTCPClient()
	tcpClient.SetAddress(":9999")
	err = tcpClient.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	time.Sleep(1 * time.Second)

	tcpClient.Stop()
	tcpServer.Stop()

	time.Sleep(1 * time.Second)
}

func ExampleNewTCPServer() {
	tcpServer := NewTCPServer()

	tcpServer.SetAddress(":9999")
	err := tcpServer.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	time.Sleep(1 * time.Second)

	tcpClient := NewTCPClient()
	tcpClient.SetAddress(":9999")
	err = tcpClient.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	time.Sleep(1 * time.Second)

	tcpClient.Stop()
	tcpServer.Stop()

	time.Sleep(1 * time.Second)

	// Output
	//
}
