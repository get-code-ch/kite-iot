package main

import (
	"fmt"
	kite "github.com/get-code-ch/kite-common"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

type Iot struct {
	conf      *IotConf
	conn      *websocket.Conn
	wg        sync.WaitGroup
	ics       map[string]*IC
	endpoints map[kite.Address]*EndpointConn
	sync      sync.Mutex
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
				go iot.provisioning(message.Data)
				break
			case kite.A_ACCEPTED:
				break
			default:
				//log.Printf("Message received -> %v", message.Data)
				break
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

			iot.sync.Lock()
			if err := action.IsValid(); err == nil {
				switch action {
				default:
					msg = string(parsed[3])
					message := kite.Message{Action: action, Sender: iot.conf.Address, Receiver: to, Data: msg}
					if err := iot.conn.WriteJSON(message); err != nil {
						iot.sync.Unlock()
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
		iot.sync.Unlock()
	}
}

func main() {
	chanMsg := make(chan []byte)

	// Loading configuration
	configFile := ""
	if len(os.Args) >= 2 {
		configFile = os.Args[1]
	}
	iot := new(Iot)
	iot.conf = loadConfig(configFile)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for {
			if iot.conn = connectServer(iot, iot.conf.Address); iot.conn == nil {
				continue
			}

			iot.wg.Add(1)

			// Listening new server message
			go iot.waitMessage()

			// Sending message
			go iot.sendMessage(chanMsg)

			iot.wg.Wait()
		}
		wg.Done()
	}()

	wg.Wait()

}
