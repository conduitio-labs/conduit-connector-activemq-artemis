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
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/go-stomp/stomp/v3"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestDestination(t *testing.T) {
	is := is.New(t)

	queueID := uuid.New().String()
	queueName := fmt.Sprintf("/queue/%s", queueID)

	ctx := context.Background()

	payloadStr := fmt.Sprintf("test message from %s", queueID)
	payload := []byte(payloadStr)

	doneCh := make(chan struct{})
	defer func() {
		timeout := time.After(5 * time.Second)
		select {
		case <-timeout:
			t.Fatal("timeout waiting for doneCh")
		case <-doneCh:
		}
	}()

	netConn, err := net.Dial("tcp", "localhost:61613")
	is.NoErr(err)

	stompConn, err := stomp.Connect(netConn, stomp.ConnOpt.Login("admin", "admin"))
	is.NoErr(err)

	sub, err := stompConn.Subscribe(queueName, stomp.AckClientIndividual)
	is.NoErr(err)

	go func() {
		msg, err := sub.Read()
		is.NoErr(err)

		var result struct {
			Payload struct {
				After sdk.RawData `json:"after"`
			}
		}
		err = json.Unmarshal(msg.Body, &result)
		is.NoErr(err)

		is.Equal(string(result.Payload.After), payloadStr)

		err = stompConn.Ack(msg)
		is.NoErr(err)

		doneCh <- struct{}{}
	}()

	config := cfgToMap(Config{
		URL:      "localhost:61613",
		User:     "admin",
		Password: "admin",
		Queue:    queueName,
	})

	destination := NewDestination()
	err = destination.Configure(ctx, config)

	err = destination.Open(ctx)
	is.NoErr(err)

	_, err = destination.Write(ctx, []sdk.Record{
		{Payload: sdk.Change{After: sdk.RawData(payload)}},
	})
	is.NoErr(err)

	err = destination.Teardown(ctx)
	is.NoErr(err)
}
