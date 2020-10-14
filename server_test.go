package net

import (
	"fmt"
	"testing"
	"time"
)

type EchoHandler struct{}

func (eh EchoHandler) OnRead(ctx *Context, in interface{}) (interface{}, error) {
	fmt.Println("Server onRead()")
	ctx.Write(in)
	return nil, nil
}

type PrintHandler struct{}

func (ph PrintHandler) OnConnect(ctx *Context) error {
	fmt.Println("Client connected")
	buffer := NewBuffer()
	buffer.Write([]byte("Hello"))
	ctx.Write(buffer)
	return nil
}

func (ph PrintHandler) OnRead(ctx *Context, in interface{}) (interface{}, error) {
	fmt.Printf("Client onRead(): %s\n", string(string(in.(*Buffer).Data())))
	in.(*Buffer).Flush()
	return nil, nil
}

func TestNewTCPServer(t *testing.T) {

	tcpServer := NewTCPServer()
	go serverProcess(tcpServer)
	time.Sleep(1 * time.Second)

	tcpClient := NewTCPClient()
	go clientProcess(tcpClient)
	time.Sleep(1 * time.Second)

	// time.Sleep(100 * time.Second)

	tcpClient.Stop()
	tcpServer.Stop()

	time.Sleep(1 * time.Second)
}

func serverProcess(server *TCPServer) {
	var echoHandler EchoHandler

	server.SetAddress(":9999")
	server.AddHandler(echoHandler)
	err := server.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	server.WaitForDone()
}

func clientProcess(client *TCPClient) {
	var printHandler PrintHandler
	client.SetAddress(":9999")
	client.AddHandler(printHandler)
	// client.pipeline().AddHandler(printHandler)
	err := client.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	client.WaitForDone()
}
