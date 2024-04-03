// Copyright © 2024 Meroxa, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package activemq

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/go-stomp/stomp/v3"
	"github.com/google/uuid"
)

func TestAcceptance(t *testing.T) {
	sourceConfig := map[string]string{
		"url":              "localhost:61613",
		"user":             "admin",
		"password":         "admin",
		"subscriptionType": "ANYCAST",

		// we want to disable artemis flow control, so that messages are delivered as soon as possible.
		// This prevents source reads from timing out in unexpected ways.
		"consumerWindowSize": "-1",
	}
	destinationConfig := map[string]string{
		"url":             "localhost:61613",
		"user":            "admin",
		"password":        "admin",
		"destinationType": "ANYCAST",
	}

	driver := sdk.ConfigurableAcceptanceTestDriver{
		Config: sdk.ConfigurableAcceptanceTestDriverConfig{
			Connector:         Connector,
			SourceConfig:      sourceConfig,
			DestinationConfig: destinationConfig,
			BeforeTest: func(t *testing.T) {
				destination := uniqueDestinationName(t)
				sourceConfig["destination"] = destination
				destinationConfig["destination"] = destination
				destinationConfig["destinationHeader"] = destination
			},
			Skip: []string{
				// Configure tests are faulty since we rely on paramgen to validate required parameters.
				"TestSource_Configure_RequiredParams",
				"TestDestination_Configure_RequiredParams",
			},
			WriteTimeout: 500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
		},
	}

	sdk.AcceptanceTest(t, driver)
}

func uniqueDestinationName(t *testing.T) string {
	return fmt.Sprintf("/queue/%s_%s", t.Name(), uuid.New().String()[0:8])
}

func TestServerTLSConfiguration(t *testing.T) {
	destination := "/queue/test"
	conn, err := connect(context.Background(), Config{
		URL:         "localhost:61617",
		User:        "admin",
		Password:    "admin",
		Destination: destination,
		TLS: TLSConfig{
			Enabled:            true,
			ClientKeyPath:      "test/certs/client_key.pem",
			ClientCertPath:     "test/certs/client_cert.pem",
			CaCertPath:         "test/certs/broker_cert",
			InsecureSkipVerify: true,
		},
	})
	if err != nil {
		log.Fatal("Failed to dial to ActiveMQ:", err)
	}
	defer conn.Disconnect()

	// Send a message to a destination
	message := "Hello, ActiveMQ Artemis with TLS!"
	err = conn.Send(destination, "text/plain", []byte(message), nil)
	if err != nil {
		log.Fatal("Failed to send message:", err)
	}
	fmt.Println("Message sent successfully!")

	// Subscribe to the destination
	sub, err := conn.Subscribe(destination, stomp.AckClient)
	if err != nil {
		log.Fatal("Failed to subscribe to destination:", err)
	}
	defer sub.Unsubscribe()

	// Receive messages from the subscription
	for {
		msg := <-sub.C
		if msg == nil {
			break
		}
		fmt.Println("Received message:", string(msg.Body))
		msg.Conn.Ack(msg)
	}

	// Wait for a short duration before disconnecting
	time.Sleep(1 * time.Second)
}
