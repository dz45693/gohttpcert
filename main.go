package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

func main() {
	go HttpServer()
	time.Sleep(1000)
	go HttpClient()
	var s string
	fmt.Scan(&s)

}

func HttpServer() {
	/*简单方式*/
	/*
		server := &http.Server{
			Addr:         ":8080",
			ReadTimeout:  5 * time.Minute, // 5 min to allow for delays when 'curl' on OSx prompts for username/password
			WriteTimeout: 10 * time.Second,
			TLSConfig:    &tls.Config{ServerName: "localhost"},
		}
	*/
	/*高级方式 使用ca.pem*/

	http.HandleFunc("/", handle)
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Minute, // 5 min to allow for delays when 'curl' on OSx prompts for username/password
		WriteTimeout: 10 * time.Second,
		TLSConfig:    getTLSConfig("localhost", "ca.pem", tls.ClientAuthType(tls.RequireAndVerifyClientCert)),
	}

	http2.ConfigureServer(server, &http2.Server{})
	if err := server.ListenAndServeTLS("server.pem", "server.key"); err != nil {
		log.Fatal(err)
	}

}

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Server Got connection: %s\r\n", r.Proto)
	if r.URL.Path == "/2nd" {
		fmt.Println("Handling 2nd")
		w.Write([]byte("Hello Again!"))
		//return
	} else {
		fmt.Println("Handling 1st")
	}

	pusher, ok := w.(http.Pusher)
	if ok {
		fmt.Println("pusher")
	}
	if !ok {
		log.Println("Can't push to client")
	} else {
		err := pusher.Push("/2nd", nil)
		if err != nil {
			log.Printf("Failed push: %v", err)
		}
	}
	w.Write([]byte("Hello"))
}
func HttpClient() {
	clientCertFile := "client.pem"
	clientKeyFile := "client.key"
	caCertFile := "ca.pem"
	var cert tls.Certificate
	var err error
	if clientCertFile != "" && clientKeyFile != "" {
		cert, err = tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			log.Fatalf("Error creating x509 keypair from client cert file %s and client key file %s", clientCertFile, clientKeyFile)
		}
	}
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		fmt.Printf("Error opening cert file %s, Error: %s", caCertFile, err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	/*http 1.1
	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		},
	}
	*/
	t := &http2.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		},
	}
	client := http.Client{Transport: t, Timeout: 15 * time.Second}
	resp, err := client.Get("https://localhost:8080/")
	if err != nil {
		fmt.Printf("Failed get: %s\r\n", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed reading response body: %s\r\n", err)
	}
	fmt.Printf("Client Got response %d: %s %s\r\n", resp.StatusCode, resp.Proto, string(body))

}
func getTLSConfig(host, caCertFile string, certOpt tls.ClientAuthType) *tls.Config {
	var caCert []byte
	var err error
	var caCertPool *x509.CertPool
	if certOpt > tls.RequestClientCert {
		caCert, err = ioutil.ReadFile(caCertFile)
		if err != nil {
			fmt.Printf("Error opening cert file %s error: %v", caCertFile, err)
		}
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}

	return &tls.Config{
		ServerName: host,
		ClientAuth: certOpt,
		ClientCAs:  caCertPool,
		MinVersion: tls.VersionTLS12, // TLS versions below 1.2 are considered insecure - see https://www.rfc-editor.org/rfc/rfc7525.txt for details
	}
}
