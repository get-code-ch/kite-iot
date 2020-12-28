package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	kite "github.com/get-code-ch/kite-common"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Iot struct {
	conf      *IotConf
	conn      *websocket.Conn
	wg        sync.WaitGroup
	ics       map[string]*IC
	endpoints []kite.Endpoint
}

func (iot *Iot) waitMessage() {
	for {
		message := kite.Message{}

		if err := iot.conn.ReadJSON(&message); err != nil {
			log.Printf("Error on readMessage -> %v", err)
			iot.wg.Done()
			return
		} else {
			switch message.Action {
			// Receiving provisioning data
			case kite.A_PROVISION:
				iot.provisioning(message.Data)
			default:
				log.Printf("Message received -> %v", message.Data)
			}
		}
	}
}

func (iot *Iot) sendMessage(input chan []byte) {
	inputRe := regexp.MustCompile(`^([^:@]*)(?:@([^:]*))?:(.+)$`)

	for {
		// Parsing input string
		if parsed := inputRe.FindSubmatch(<-input); parsed != nil {
			to := kite.Address{Domain: "*", Type: kite.H_ANY, Host: "*", Address: "*", Id: "*"}
			msg := ""

			action := kite.Action(strings.ToLower(string(parsed[1])))
			to.StringToAddress(string(parsed[2]))

			if err := action.IsValid(); err == nil {
				switch action {
				default:
					msg = string(parsed[3])
					message := kite.Message{Action: action, Sender: iot.conf.Address, Receiver: to, Data: msg}
					if err := iot.conn.WriteJSON(message); err != nil {
						return
					}
				}
			} else {
				log.Printf("%s", err)
				fmt.Printf("%s> ", iot.conf.Address)
			}
		} else {
			log.Printf("Invalid command ({action}[@{destination}]{:{message}})")
			fmt.Printf("%s> ", iot.conf.Address)
		}
	}
}

/*
func (iot *Iot) readStdin(input chan []byte) {
	for {
		fmt.Printf("%s> ", iot.conf.Endpoint)
		msg := bufio.NewScanner(os.Stdin)
		msg.Scan()
		if len(msg.Bytes()) == 0 {
			iot.wg.Done()
			return
		}
		input <- msg.Bytes()
	}
}
*/

func main() {
	var response *http.Response
	var err error

	chanMsg := make(chan []byte)

	// Loading configuration
	configFile := ""
	if len(os.Args) >= 2 {
		configFile = os.Args[1]
	}
	iot := new(Iot)
	iot.conf = loadConfig(configFile)
	// Configure Server URL
	addr := flag.String("addr", fmt.Sprintf("%s:%s", iot.conf.Server, iot.conf.Port), "kite server http(s) address")
	flag.Parse()

	serverURL := url.URL{}
	if iot.conf.Ssl {
		serverURL = url.URL{Scheme: "wss", Host: *addr, Path: "/ws"}
	} else {
		serverURL = url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	}

	// Adding origin in header, for server cross origin resource sharing (CORS) check
	header := http.Header{}
	header.Set("Origin", serverURL.String())

	// Connecting kite server, if connection failed retrying every x seconds
	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for {
			if iot.conn, response, err = dialer.Dial(serverURL.String(), header); err != nil {
				iot.conn = nil
				log.Printf("Dial Error %v next try in 5 seconds\n", err)

				for {
					time.Sleep(5 * time.Second)
					if iot.conn, response, err = dialer.Dial(serverURL.String(), header); err == nil {
						break
					} else {
						log.Printf("Dial Error %v next try in 5 seconds\n", err)
					}
				}
			}
			log.Printf("kite server connectected, (http status %d)", response.StatusCode)

			// Configuring ping handler (just logging a ping on stdin
			iot.conn.SetPingHandler(func(data string) error {
				log.Printf("ping received\n")
				return nil
			})

			// Connection is now established, now we sending iot registration to server
			msg := kite.Message{Action: "register", Sender: iot.conf.Address, Data: iot.conf.ApiKey}
			if err := iot.conn.WriteJSON(msg); err != nil {
				log.Printf("Error registring iot on sever --> %v", err)
				time.Sleep(5 * time.Second)
				iot.conn = nil
				continue
			}

			// Reading server response
			if err = iot.conn.ReadJSON(&msg); err != nil {
				log.Printf("Error registring iot on sever --> %v", err)
				time.Sleep(5 * time.Second)
				iot.conn = nil
				continue
			} else {
				//TODO: Checking if returned message is an ACCEPT
				fmt.Println()
				log.Printf("Message received from %v\n", msg)
			}

			iot.wg.Add(1)

			// Listening new server message
			go iot.waitMessage()

			// Reading prompt
			//go iot.readStdin(chanMsg)

			// Sending message
			go iot.sendMessage(chanMsg)

			iot.wg.Wait()

		}
		wg.Done()
	}()

	wg.Wait()

}
