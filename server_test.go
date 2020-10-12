package net

import (
	"fmt"
	"testing"
	"time"
)

type EchoHandler struct{}

func (eh *EchoHandler) Inbound(ctx *Context, in interface{}) interface{} {
	buf := make([]byte, 1024)
	n, _ := in.(*Buffer).Read(buf)
	ctx.Write(buf[:n])
	return nil
}

func (eh *EchoHandler) Outbound(ctx *Context, in interface{}) interface{} {
	return in
}

func TestNewTCPServer(t *testing.T) {

	tcpServer := NewTCPServer()
	go serverProcess(tcpServer)
	time.Sleep(1 * time.Second)

	tcpClient := NewTCPClient()
	go clientProcess(tcpClient)
	time.Sleep(1 * time.Second)

	tcpServer.WaitForDone()

	time.Sleep(1 * time.Second)

	tcpClient.Stop()
	tcpServer.Stop()

	time.Sleep(1 * time.Second)
}

func serverProcess(server *TCPServer) {
	server.SetAddress(":9999")
	err := server.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	server.WaitForDone()
}

func clientProcess(client *TCPClient) {
	client.SetAddress(":9999")
	err := client.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	client.WaitForDone()
}
