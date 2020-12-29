package main

import (
	"crypto/tls"
	"fmt"
	kite "github.com/get-code-ch/kite-common"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"time"
)

func connectServer(iot *Iot, me kite.Address) *websocket.Conn {

	var response *http.Response
	var conn *websocket.Conn
	var err error

	serverURL := url.URL{}
	host := fmt.Sprintf("%s:%s", iot.conf.Server, iot.conf.Port)
	if iot.conf.Ssl {
		serverURL = url.URL{Scheme: "wss", Host: host, Path: "/ws"}
	} else {
		serverURL = url.URL{Scheme: "ws", Host: host, Path: "/ws"}
	}


	// Adding origin in header, for server cross origin resource sharing (CORS) check
	header := http.Header{}
	header.Set("Origin", serverURL.String())

	// Connecting kite server, if connection failed retrying every x seconds
	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	if conn, response, err = dialer.Dial(serverURL.String(), header); err != nil {
		conn = nil
		log.Printf("Dial Error %v next try in 5 seconds\n", err)

		for {
			time.Sleep(5 * time.Second)
			if conn, response, err = dialer.Dial(serverURL.String(), header); err == nil {
				break
			} else {
				log.Printf("Dial Error %v next try in 5 seconds\n", err)
			}
		}
	}
	log.Printf("kite server connectected, (http status %d)", response.StatusCode)

	// Configuring ping handler (just logging a ping on stdin
	conn.SetPingHandler(func(data string) error {
		log.Printf("ping received\n")
		return nil
	})

	// Connection is now established, now we sending iot registration to server
	message:= kite.Message{Action: kite.A_REGISTER, Sender: me, Data: iot.conf.ApiKey}
	if err := conn.WriteJSON(message); err != nil {
		log.Printf("Error registring iot on sever --> %v", err)
		time.Sleep(5 * time.Second)
		return nil
	}

	// Reading server response
	if err = conn.ReadJSON(&message); err != nil {
		log.Printf("Error registring iot on sever --> %v", err)
		time.Sleep(5 * time.Second)
		return nil
	} else {
		if message.Action == kite.A_ACCEPTED {
			log.Printf("Connection accepted from %s", message.Sender)
		} else {
			log.Printf("Unattended response from %s", message.Sender)
		}
	}
	return conn
}
