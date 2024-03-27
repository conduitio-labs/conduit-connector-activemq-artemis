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
	"fmt"
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/go-stomp/stomp/v3"
	"github.com/go-stomp/stomp/v3/frame"
	"github.com/matryer/is"
)

func TestAcceptance(t *testing.T) {
	cfg := map[string]string{
		"url":      "localhost:61613",
		"user":     "admin",
		"password": "admin",
	}

	driver := sdk.ConfigurableAcceptanceTestDriver{
		Config: sdk.ConfigurableAcceptanceTestDriverConfig{
			Connector:         Connector,
			SourceConfig:      cfg,
			DestinationConfig: cfg,
			BeforeTest: func(t *testing.T) {
				cfg["queue"] = uniqueQueueName(t)
			},

			Skip: []string{
				"TestSource_Configure_RequiredParams",
				"TestDestination_Configure_RequiredParams",
			},
			WriteTimeout: 500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
		},
	}

	sdk.AcceptanceTest(t, testDriver{driver})
}

type testDriver struct {
	sdk.ConfigurableAcceptanceTestDriver
}

func (d testDriver) GenerateRecord(t *testing.T, op sdk.Operation) sdk.Record {
	position := sdk.Position(randBytes())
	key := sdk.RawData(fmt.Sprintf("key-%s", randString()))
	payload := sdk.RawData(fmt.Sprintf("data-%s", randString()))

	return sdk.Util.Source.NewRecordCreate(position, sdk.Metadata{}, key, payload)
}
}

func teardownResource(is *is.I, res any) {
	type unsubscriber interface{ Unsubscribe() error }
	type disconnecter interface{ Disconnect() error }

	switch r := res.(type) {
	case unsubscriber:
		err := r.Unsubscribe()
		is.NoErr(err)
	case disconnecter:
		err := r.Disconnect()
		is.NoErr(err)
	}
}

func TestRawReadWrite(t *testing.T) {
	t.Run("subscribing then writing", subscribingThenWriting)
	t.Run("subscribing then writing with 2 conns", subscribingThenWriting2Conns)
	t.Run("writing then subscribing", writingThenSubscribing)
	t.Run("writing then subscribing with 2 conns", writingThenSubscribing2Conns)
}

func subscribingThenWriting(t *testing.T) {
	is := is.New(t)
	queue := uniqueQueueName(t)

	connOpts := []func(*stomp.Conn) error{
		stomp.ConnOpt.Login("admin", "admin"),
	}
	sendOpts := []func(*frame.Frame) error{
		stomp.SendOpt.Receipt,
		stomp.SendOpt.Header("destination-type", "ANYCAST"),
		stomp.SendOpt.Header("destination", queue),
	}

	conn, err := stomp.Dial("tcp", "localhost:61613", connOpts...)
	is.NoErr(err)
	defer teardownResource(is, conn)

	// Subscribing

	subscriber, err := conn.Subscribe(queue,
		stomp.AckClientIndividual,
		stomp.SubscribeOpt.Header("subscription-type", "ANYCAST"),
		stomp.SubscribeOpt.Header("destination", queue),
	)
	is.NoErr(err)
	defer teardownResource(is, subscriber)

	randMsg := randString()

	// Writing

	err = conn.Send(queue, "text/plain", []byte(randMsg), sendOpts...)
	is.NoErr(err)

	select {
	case msg := <-subscriber.C:
		is.Equal(string(msg.Body), randMsg)

		err = conn.Ack(msg)
		is.NoErr(err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func writingThenSubscribing(t *testing.T) {
	is := is.New(t)
	queue := uniqueQueueName(t)
	connOpts := []func(*stomp.Conn) error{
		stomp.ConnOpt.Login("admin", "admin"),
	}
	sendOpts := []func(*frame.Frame) error{
		stomp.SendOpt.Receipt,
		stomp.SendOpt.Header("destination-type", "ANYCAST"),
		stomp.SendOpt.Header("destination", queue),
	}

	connWriter, err := stomp.Dial("tcp", "localhost:61613", connOpts...)
	is.NoErr(err)
	defer connWriter.Disconnect()

	randMsg := randString()

	// Writing

	err = connWriter.Send(queue, "text/plain", []byte(randMsg), sendOpts...)
	is.NoErr(err)

	connReader, err := stomp.Dial("tcp", "localhost:61613", connOpts...)
	is.NoErr(err)
	defer connReader.Disconnect()

	// Subscribing

	subscriber, err := connReader.Subscribe(queue,
		stomp.AckClientIndividual,
	)
	is.NoErr(err)
	defer subscriber.Unsubscribe()

	select {
	case msg := <-subscriber.C:
		is.Equal(string(msg.Body), randMsg)
		err = connWriter.Ack(msg)
		is.NoErr(err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func subscribingThenWriting2Conns(t *testing.T) {
	is := is.New(t)
	queue := uniqueQueueName(t)
	// queue := "/queue/example-2710f86d"
	fmt.Println(queue)

	connReader, err := stomp.Dial("tcp", "localhost:61613", stomp.ConnOpt.Login("admin", "admin"))
	is.NoErr(err)
	defer connReader.Disconnect()

	// Subscribing

	subscriber, err := connReader.Subscribe(queue, stomp.AckClientIndividual,
		stomp.SubscribeOpt.Header("subscription-type", "ANYCAST"),
		stomp.SubscribeOpt.Header("destination", queue),
	)
	is.NoErr(err)
	defer subscriber.Unsubscribe()

	connWriter, err := stomp.Dial("tcp", "localhost:61613",
		stomp.ConnOpt.Login("admin", "admin"),
	)
	is.NoErr(err)
	defer connWriter.Disconnect()

	// Writing

	randMsg := randString()
	err = connWriter.Send(queue, "text/plain", []byte(randMsg),
		stomp.SendOpt.Receipt,
		stomp.SendOpt.Header("destination-type", "ANYCAST"),
		stomp.SendOpt.Header("destination", queue),
	)
	is.NoErr(err)

	select {
	case msg := <-subscriber.C:
		// is.Equal(string(msg.Body), randMsg)

		err = connReader.Ack(msg)
		is.NoErr(err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func writingThenSubscribing2Conns(t *testing.T) {
	is := is.New(t)
	queue := uniqueQueueName(t)

	connOpts := []func(*stomp.Conn) error{
		stomp.ConnOpt.Login("admin", "admin"),
	}
	sendOpts := []func(*frame.Frame) error{
		stomp.SendOpt.Receipt,
		stomp.SendOpt.Header("destination-type", "ANYCAST"),
		stomp.SendOpt.Header("destination", queue),
	}

	connWriter, err := stomp.Dial("tcp", "localhost:61613", connOpts...)
	is.NoErr(err)
	defer connWriter.Disconnect()

	randMsg := randString()

	// Writing

	err = connWriter.Send(queue, "text/plain", []byte(randMsg), sendOpts...)
	is.NoErr(err)

	connReader, err := stomp.Dial("tcp", "localhost:61613", connOpts...)
	is.NoErr(err)
	defer connReader.Disconnect()

	// Subscribing

	subscriber, err := connReader.Subscribe(queue, stomp.AckClientIndividual)
	is.NoErr(err)
	defer subscriber.Unsubscribe()

	select {
	case msg := <-subscriber.C:
		is.Equal(string(msg.Body), randMsg)

		err = connReader.Ack(msg)
		is.NoErr(err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}
