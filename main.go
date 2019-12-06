package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Message struct {
	Wisdom string
	Secret string
	Team   string
}

type Result struct {
	One   string
	Two   string
	Three string
	Four  string
}

type Payload struct {
	Time   time.Time `json:"time"`
	Wisdom string    `json:"wisdom"`
	Secret string    `json:"secret"`
	Team   string    `json:"team"`
}

func parseMessage(data *Message, result *Result) {
	switch data.Wisdom[:1] {
	case "1":
		result.One = data.Wisdom
	case "2":
		result.Two = data.Wisdom
	case "3":
		result.Three = data.Wisdom
	case "4":
		result.Four = data.Wisdom
	}
}

func checkResult(result *Result) bool {
	if strings.HasPrefix(result.One, "1") &&
		strings.HasPrefix(result.Two, "2") &&
		strings.HasPrefix(result.Three, "3") &&
		strings.HasPrefix(result.Four, "4") {
		return true
	}
	return false
}

func hashResponse(data *Message, result *Result) []byte {
	var payload Payload
	payload.Wisdom = result.One[3:] + result.Two[2:] + result.Three[2:] + result.Four[2:]
	payload.Secret = hex.EncodeToString(sha1.New().Sum([]byte(data.Secret + "Za_stepuhu + Vlad")))
	payload.Time = time.Now()
	payload.Team = "Za_stepuhu + Vlad"
	hashed, _ := json.Marshal(&payload)
	return hashed
}

func main() {
	// rutine for parallel handling
	done := make(chan bool)
	// setting options
	host := "localhost"
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:1883", host))
	// creating instance of client
	client := mqtt.NewClient(opts)
	// awaiting for it connection
	client.Connect().Wait()
	// creating instance of final message
	var result Result
	go func() {
		// subscription
		client.Subscribe("/test/inception", 0, func(client mqtt.Client, message mqtt.Message) {
			// struct for parsing for message
			var data Message
			// from json to struct
			if err := json.Unmarshal(message.Payload(), &data); err != nil {
				return
			}
			fmt.Printf(data.Wisdom)
			fmt.Printf("\n")
			// parse message to special struct
			parseMessage(&data, &result)
			// if we have enough messages we send reponse
			if checkResult(&result) {
				// hashing answer
				hashed := hashResponse(&data, &result)
				// sending answer
				client.Publish("/test/result", 0, false, hashed)
				done <- true
			}
		})
	}()
	<-done // Awaiting for all requests
}
