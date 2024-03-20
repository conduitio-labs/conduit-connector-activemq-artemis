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
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/google/uuid"
	"github.com/matryer/is"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
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
	return sdk.Record{
		Position:  sdk.Position(randString()), // position doesn't matter, as long as it's unique
		Operation: op,
		Metadata:  map[string]string{randString(): randString()},
		Key:       sdk.RawData(fmt.Sprintf("key-%s", randString())),
		Payload: sdk.Change{
			Before: nil,
			After:  sdk.RawData(fmt.Sprintf("payload-%s", randString())),
		},
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func uniqueQueueName(t *testing.T) string {
	return fmt.Sprintf("%s_%s", t.Name(), uuid.New().String())
}

func init() {
	log := zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.DefaultContextLogger = &log
}

func TestDestination_Write_Success(t *testing.T) {
	is := is.New(t)
	ctx := context.Background()

	dest, cleanup := openDestination(ctx, t)
	defer cleanup()

	want := generateRecords(t, sdk.OperationCreate, 20)

	n, err := dest.Write(ctx, want)
	is.NoErr(err)
	is.Equal(n, len(want))

	ReadFromDestination(t, want)
}

func openDestination(ctx context.Context, t *testing.T) (dest sdk.Destination, cleanup func()) {
	is := is.New(t)

	dest = NewDestination()
	err := dest.Configure(ctx, config)
	is.NoErr(err)

	openCtx, cancelOpenCtx := context.WithCancel(ctx)
	err = dest.Open(openCtx)
	is.NoErr(err)

	// make sure connector is cleaned up only once
	var cleanupOnce sync.Once
	cleanup = func() {
		cleanupOnce.Do(func() {
			cancelOpenCtx()
			err = dest.Teardown(ctx)
			is.NoErr(err)
		})
	}

	return dest, cleanup
}

var config = map[string]string{
	"url":      "localhost:61613",
	"user":     "admin",
	"password": "admin",
	"queue":    "scripts",
}

func generateRecords(t *testing.T, op sdk.Operation, count int) []sdk.Record {
	records := make([]sdk.Record, count)
	for i := range records {
		records[i] = GenerateRecord(t, op)
	}
	return records
}

func GenerateRecord(t *testing.T, op sdk.Operation) sdk.Record {
	return sdk.Record{
		Operation: op,
		Payload: sdk.Change{
			Before: nil,
			After:  sdk.RawData("payload"),
		},
	}
}

func ReadFromDestination(t *testing.T, records []sdk.Record) []sdk.Record {

	is := is.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// writing something to the destination should result in the same record
	// being produced by the source
	src := NewSource()
	err := src.Configure(ctx, config)
	is.NoErr(err)

	err = src.Open(ctx, nil)
	is.NoErr(err)

	output := make([]sdk.Record, 0, len(records))
	for i := 0; i < cap(output); i++ {
		r, err := src.Read(ctx)
		is.NoErr(err)
		output = append(output, r)
	}

	readCtx, readCancel := context.WithTimeout(ctx, time.Second)
	defer readCancel()
	r, err := src.Read(readCtx)
	is.Equal(sdk.Record{}, r) // record should be empty
	is.True(errors.Is(err, context.DeadlineExceeded) || errors.Is(err, sdk.ErrBackoffRetry))

	return output
}
