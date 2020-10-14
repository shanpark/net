package net

import (
	"fmt"
	"testing"
	"time"
)

type EchoHandler struct{}

func (eh EchoHandler) OnConnect(ctx *Context) error {
	fmt.Println("Server OnConnect()")
	return nil
}

func (eh EchoHandler) OnRead(ctx *Context, in interface{}) (interface{}, error) {
	fmt.Println("Server OnRead()")
	ctx.Write(in)
	return nil, nil
}

func (eh EchoHandler) OnWrite(ctx *Context, out interface{}) (interface{}, error) {
	fmt.Println("Server OnWrite()")
	return out, nil
}

func (eh EchoHandler) OnDisconnect(ctx *Context) {
	fmt.Println("Server OnDisconnect()")
}

func (eh EchoHandler) OnError(ctx *Context, err error) {
	fmt.Printf("Error! - %v\n", err)
}

type PrintHandler struct{}

func (ph PrintHandler) OnConnect(ctx *Context) error {
	fmt.Println("Client OnConnect")
	buffer := NewBuffer()
	buffer.Write([]byte("Hello"))
	ctx.Write(buffer)
	return nil
}

func (ph PrintHandler) OnRead(ctx *Context, in interface{}) (interface{}, error) {
	fmt.Printf("Client OnRead: %s\n", string(string(in.(*Buffer).Data())))
	ctx.Close()
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
}

func serverProcess(server *TCPServer) {
	server.SetAddress(":9999")
	server.AddHandler(EchoHandler{})
	err := server.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	server.WaitForDone()
}

func clientProcess(client *TCPClient) {
	client.SetAddress(":9999")
	client.AddHandler(PrintHandler{})
	err := client.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	client.WaitForDone()
}
