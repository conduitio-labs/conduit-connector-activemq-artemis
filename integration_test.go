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
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestConnectorSingleReadWrite(t *testing.T) {
	is := is.New(t)

	queueID := uuid.New().String()
	artemisDestination := fmt.Sprintf("/queue/%s", queueID)

	config := Config{
		URL:      "localhost:61613",
		User:     "admin",
		Password: "admin",
		Queue:    artemisDestination,
	}

	ctx := context.Background()

	payloadStr := fmt.Sprintf("test message from %s", queueID)
	payload := []byte(payloadStr)

	doneCh := make(chan struct{})
	defer func() {
		timeout := time.After(3 * time.Second)
		select {
		case <-timeout:
			t.Fatal("timeout waiting for doneCh")
		case <-doneCh:
		}
	}()

	configMap := cfgToMap(config)

	source := NewSource()
	err := source.Configure(ctx, configMap)
	is.NoErr(err)

	err = source.Open(ctx, nil)
	is.NoErr(err)

	go func() {
		rec, err := source.Read(ctx)
		is.NoErr(err)

		var result struct {
			Payload struct {
				After sdk.RawData `json:"after"`
			}
		}
		err = json.Unmarshal(rec.Payload.After.Bytes(), &result)
		is.NoErr(err)

		recvPayload := string(result.Payload.After.Bytes())

		is.Equal(recvPayload, payloadStr)

		doneCh <- struct{}{}
	}()

	destination := NewDestination()
	err = destination.Configure(ctx, cfgToMap(config))

	err = destination.Open(ctx)
	is.NoErr(err)

	totalWritten, err := destination.Write(ctx, []sdk.Record{
		{Payload: sdk.Change{After: sdk.RawData(payload)}},
	})
	is.Equal(totalWritten, 1)

	is.NoErr(err)

	err = destination.Teardown(ctx)
	is.NoErr(err)
}

func TestConnectorRestartFull(t *testing.T) {
	is := is.New(t)

	queueID := uuid.New().String()
	artemisDestination := fmt.Sprintf("/queue/%s", queueID)
	cfgMap := cfgToMap(Config{
		URL:      "localhost:61613",
		User:     "admin",
		Password: "admin",
		Queue:    artemisDestination,
	})

	recs1 := generateSDKRecords(1, 6)

	produce(is, cfgMap, recs1)
	lastPosition := testSourceIntegrationRead(t, cfgMap, nil, recs1, false)

	recs2 := generateSDKRecords(4, 6)
	produce(is, cfgMap, recs2)
	testSourceIntegrationRead(t, cfgMap, lastPosition, recs2, false)
}

func testSourceIntegrationRead(
	t *testing.T,
	cfgMap map[string]string,
	startFrom sdk.Position,
	wantRecords []sdk.Record,
	ackFirstOnly bool,
) sdk.Position {
	is := is.New(t)
	ctx := context.Background()

	underTest := NewSource()
	defer func() {
		err := underTest.Teardown(ctx)
		is.NoErr(err)
	}()

	err := underTest.Configure(ctx, cfgMap)
	is.NoErr(err)
	err = underTest.Open(ctx, startFrom)
	is.NoErr(err)

	var positions []sdk.Position
	for _, wantRecord := range wantRecords {
		rec, err := underTest.Read(ctx)
		is.NoErr(err)
		is.Equal(wantRecord.Key, rec.Key.Bytes())

		positions = append(positions, rec.Position)
	}

	for i, p := range positions {
		if i > 0 && ackFirstOnly {
			break
		}
		err = underTest.Ack(ctx, p)
		is.NoErr(err)
	}

	return positions[len(positions)-1]
}
