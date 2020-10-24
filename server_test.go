package net

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"
)

type EchoHandler struct{}

func (eh EchoHandler) OnConnect(ctx *TCPContext) error {
	fmt.Println("Server OnConnect()")
	return nil
}

func (eh EchoHandler) OnRead(ctx *TCPContext, in interface{}) (interface{}, error) {
	fmt.Println("Server OnRead()")
	ctx.Write(in)
	return nil, nil
}

func (eh EchoHandler) OnWrite(ctx *TCPContext, out interface{}) (interface{}, error) {
	fmt.Println("Server OnWrite()")
	return out, nil
}

func (eh EchoHandler) OnDisconnect(ctx *TCPContext) {
	fmt.Println("Server OnDisconnect()")
}

func (eh EchoHandler) OnError(ctx *TCPContext, err error) {
	fmt.Printf("Error! - %v\n", err)
	ctx.Close()
}

type PrintHandler struct{}

func (ph PrintHandler) OnConnect(ctx *TCPContext) error {
	fmt.Println("Client OnConnect")
	buffer := NewBuffer(256)
	buffer.Write([]byte("Hello"))
	ctx.Write(buffer)
	return nil
}

func (ph PrintHandler) OnRead(ctx *TCPContext, in interface{}) (interface{}, error) {
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

type HTTPHandler struct{}

func (ph HTTPHandler) OnConnect(ctx *TCPContext) error {
	fmt.Println("Client OnConnect:")
	buffer := NewBuffer(256)
	buffer.Write([]byte("GET / HTTP/1.1\n\n"))
	if ctx.Write(buffer) != nil {
		fmt.Println("ctx.Write(buffer) error !!")
	}
	return nil
}

func (ph HTTPHandler) OnRead(ctx *TCPContext, in interface{}) (interface{}, error) {
	data := in.(*Buffer).Data()
	fmt.Printf("Client OnRead: %d\n", len(data))
	fmt.Println(string(data))
	ctx.Close()
	return nil, nil
}

func (ph HTTPHandler) OnError(ctx *TCPContext, err error) {
	fmt.Println("Client OnError: ", err)
	ctx.Close()
}

func (ph HTTPHandler) OnDisconnect(ctx *TCPContext) {
	fmt.Println("Client OnDisconnect")
}

func ExampleNewTCPClient() {
	tcpClient := NewTCPClient()
	tcpClient.SetAddress("www.google.com:80")
	tcpClient.AddHandler(HTTPHandler{})
	tcpClient.Start()
	tcpClient.WaitForDone()
	fmt.Println("stopped.")

	time.Sleep(1 * time.Second)

	// Output:
	// stopped.
}

func ExampleNewTLSClient() {
	// config := &tls.Config{InsecureSkipVerify: true}
	config := &tls.Config{ServerName: "google.com"}
	tlsClient := NewTLSClient(config)
	tlsClient.SetAddress("www.google.com:443")
	tlsClient.AddHandler(HTTPHandler{})
	tlsClient.Start()
	tlsClient.WaitForDone()
	fmt.Println("stopped.")

	time.Sleep(1 * time.Second)

	// Output:
	// stopped.
}

type ToHandler struct{}

func (ph ToHandler) OnTimeout(ctx *TCPContext) error {
	fmt.Printf("Client OnTimeout:\n")
	ctx.Close()
	return nil
}

func ExampleNewTCPClient_timeout() {
	tcpClient := NewTCPClient()
	tcpClient.SetAddress("192.168.1.123:80")
	tcpClient.SetTimeout(1*time.Second, 1*time.Second)
	tcpClient.AddHandler(ToHandler{})
	if err := tcpClient.Start(); err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	tcpClient.WaitForDone()

	fmt.Println("stopped.")

	// Output:
	// Client OnTimeout:
	// stopped.
}
