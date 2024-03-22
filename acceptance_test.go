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
	"os"
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/matryer/is"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
			WriteTimeout: 5000 * time.Millisecond,
			ReadTimeout:  5000 * time.Millisecond,
		},
	}

	// sdk.AcceptanceTest(t, driver)
	sdk.AcceptanceTest(t, testDriver{driver})
}

type testDriver struct {
	sdk.ConfigurableAcceptanceTestDriver
}

func (d testDriver) GenerateRecord(t *testing.T, op sdk.Operation) sdk.Record {
	return sdk.Record{
		Position:  sdk.Position(randString()),
		Operation: op,
		Metadata:  map[string]string{randString(): randString()},
		Key:       sdk.RawData(fmt.Sprintf("key-%s", randString())),
		Payload: sdk.Change{
			After: sdk.RawData(fmt.Sprintf("payload-%s", randString())),
		},
	}
}

func init() {
	log := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.DefaultContextLogger = &log
}

func TestWriteRead20Recs(t *testing.T) {
	is := is.New(t)
	cfg := map[string]string{
		"url":      "localhost:61613",
		"user":     "admin",
		"password": "admin",
		// "queue":    "acceptance_queue",
	}
	cfg["queue"] = uniqueQueueName(t)

	driver := sdk.ConfigurableAcceptanceTestDriver{
		Config: sdk.ConfigurableAcceptanceTestDriverConfig{
			Connector:         Connector,
			SourceConfig:      cfg,
			DestinationConfig: cfg,
			WriteTimeout:      5000 * time.Millisecond,
			ReadTimeout:       5000 * time.Millisecond,
		},
	}
	_ = driver

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	destination := NewDestination()

	err := destination.Configure(context.Background(), cfg)
	is.NoErr(err)

	err = destination.Open(ctx)
	is.NoErr(err)
	defer destination.Teardown(ctx)

	total := 20

	for i := 0; i < total; i++ {
		rec := driver.GenerateRecord(t, sdk.OperationCreate)

		_, err := destination.Write(ctx, []sdk.Record{rec})
		is.NoErr(err)
	}

	source := NewSource()
	err = source.Configure(ctx, cfg)
	is.NoErr(err)
	err = source.Open(ctx, nil)
	is.NoErr(err)
	defer source.Teardown(ctx)

	for i := 0; i < total; i++ {
		rec, err := source.Read(ctx)
		is.NoErr(err) // Failed to read from source

		err = source.Ack(ctx, rec.Position)
		is.NoErr(err) // Failed to ack source
	}
}
