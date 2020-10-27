package net

import (
	"crypto/tls"
	"fmt"
	"log"
	"testing"
	"time"
)

const serverKey = `-----BEGIN EC PARAMETERS-----
BggqhkjOPQMBBw==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHg+g2unjA5BkDtXSN9ShN7kbPlbCcqcYdDu+QeV8XWuoAoGCCqGSM49
AwEHoUQDQgAEcZpodWh3SEs5Hh3rrEiu1LZOYSaNIWO34MgRxvqwz1FMpLxNlx0G
cSqrxhPubawptX5MSr02ft32kfOlYbaF5Q==
-----END EC PRIVATE KEY-----
`

const serverCert = `-----BEGIN CERTIFICATE-----
MIIB+TCCAZ+gAwIBAgIJAL05LKXo6PrrMAoGCCqGSM49BAMCMFkxCzAJBgNVBAYT
AkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRn
aXRzIFB0eSBMdGQxEjAQBgNVBAMMCWxvY2FsaG9zdDAeFw0xNTEyMDgxNDAxMTNa
Fw0yNTEyMDUxNDAxMTNaMFkxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0
YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxEjAQBgNVBAMM
CWxvY2FsaG9zdDBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABHGaaHVod0hLOR4d
66xIrtS2TmEmjSFjt+DIEcb6sM9RTKS8TZcdBnEqq8YT7m2sKbV+TEq9Nn7d9pHz
pWG2heWjUDBOMB0GA1UdDgQWBBR0fqrecDJ44D/fiYJiOeBzfoqEijAfBgNVHSME
GDAWgBR0fqrecDJ44D/fiYJiOeBzfoqEijAMBgNVHRMEBTADAQH/MAoGCCqGSM49
BAMCA0gAMEUCIEKzVMF3JqjQjuM2rX7Rx8hancI5KJhwfeKu1xbyR7XaAiEA2UT7
1xOP035EcraRmWPe7tO0LpXgMxlh2VItpc2uc2w=
-----END CERTIFICATE-----
`

type EchoHandler struct{}

func (eh EchoHandler) OnConnect(ctx *SoContext) error {
	fmt.Println("Server OnConnect()")
	return nil
}

func (eh EchoHandler) OnRead(ctx *SoContext, in interface{}) (interface{}, error) {
	fmt.Printf("Server OnRead(): %s\n", string(string(in.(*Buffer).Data())))
	ctx.Write(in)
	return nil, nil
}

func (eh EchoHandler) OnWrite(ctx *SoContext, out interface{}) (interface{}, error) {
	fmt.Println("Server OnWrite()")
	return out, nil
}

func (eh EchoHandler) OnDisconnect(ctx *SoContext) {
	fmt.Println("Server OnDisconnect()")
}

func (eh EchoHandler) OnError(ctx *SoContext, err error) {
	fmt.Printf("Server Error! - %v\n", err)
	ctx.Close()
}

type PrintHandler struct{}

func (ph PrintHandler) OnConnect(ctx *SoContext) error {
	fmt.Println("Client OnConnect")
	buffer := NewBuffer(256)
	buffer.Write([]byte("Hello"))
	ctx.Write(buffer)
	return nil
}

func (ph PrintHandler) OnRead(ctx *SoContext, in interface{}) (interface{}, error) {
	fmt.Printf("Client OnRead: %s\n", string(string(in.(*Buffer).Data())))
	ctx.Close()
	return nil, nil
}

func (ph PrintHandler) OnWrite(ctx *SoContext, out interface{}) (interface{}, error) {
	fmt.Println("Client OnWrite")
	return out, nil
}

func (ph PrintHandler) OnDisconnect(ctx *SoContext) {
	fmt.Println("Client OnDisconnect")
}

func (ph PrintHandler) OnError(ctx *SoContext, err error) {
	fmt.Printf("Client Error! - %v\n", err)
	ctx.Close()
}

func TestNewTLSServer(t *testing.T) {
	cer, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		log.Fatal(err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	tlsServer := NewTLSServer(config)
	go serverProcess(tlsServer)
	time.Sleep(1 * time.Second)

	config = &tls.Config{InsecureSkipVerify: true}
	tlsClient := NewTLSClient(config)
	go clientProcess(tlsClient)
	time.Sleep(1 * time.Second)

	// time.Sleep(100 * time.Second)

	tlsClient.Stop()
	tlsServer.Stop()
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

func serverProcess(server SoService) {
	server.SetAddress(":9999")
	server.AddHandler(EchoHandler{})
	err := server.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	server.WaitForDone()
}

func clientProcess(client SoService) {
	client.SetAddress(":9999")
	client.AddHandler(PrintHandler{})
	err := client.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	client.WaitForDone()
}

type HTTPHandler struct{}

func (ph HTTPHandler) OnConnect(ctx *SoContext) error {
	fmt.Println("Client OnConnect:")
	buffer := NewBuffer(256)
	buffer.Write([]byte("GET / HTTP/1.1\n\n"))
	if ctx.Write(buffer) != nil {
		fmt.Println("ctx.Write(buffer) error !!")
	}
	return nil
}

func (ph HTTPHandler) OnRead(ctx *SoContext, in interface{}) (interface{}, error) {
	data := in.(*Buffer).Data()
	fmt.Printf("Client OnRead: %d\n", len(data))
	fmt.Println(string(data))
	ctx.Close()
	return nil, nil
}

func (ph HTTPHandler) OnError(ctx *SoContext, err error) {
	fmt.Println("Client OnError: ", err)
	ctx.Close()
}

func (ph HTTPHandler) OnDisconnect(ctx *SoContext) {
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

func (ph ToHandler) OnTimeout(ctx *SoContext) error {
	fmt.Printf("Client OnTimeout:\n")
	ctx.Close()
	return nil
}

func ExampleNewTCPClient_timeout() {
	tcpClient := NewTCPClient()
	// tcpClient.SetAddress("192.168.1.123:80")
	tcpClient.SetAddress("www.naver.com:80")
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
