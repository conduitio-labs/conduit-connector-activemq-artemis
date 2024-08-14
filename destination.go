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

	"github.com/conduitio/conduit-commons/config"
	"github.com/conduitio/conduit-commons/opencdc"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/go-stomp/stomp/v3"
)

type DestinationConfig struct {
	Config

	// DestinationType is the routing type of the destination. It can be either
	// ANYCAST or MULTICAST, with ANYCAST being the default.
	// Maps to the "destination-type" header in the STOMP SEND frame.
	DestinationType string `json:"destinationType" default:"ANYCAST" validation:"inclusion=ANYCAST|MULTICAST"`

	// DestinationHeader maps to the "destination" header in the STOMP SEND
	// frame. Useful when using ANYCAST.
	DestinationHeader string `json:"destinationHeader"`
}

type Destination struct {
	sdk.UnimplementedDestination
	config DestinationConfig

	conn *stomp.Conn
}

func NewDestination() sdk.Destination {
	return sdk.DestinationWithMiddleware(&Destination{}, sdk.DefaultDestinationMiddleware()...)
}

func (d *Destination) Parameters() config.Parameters {
	return d.config.Parameters()
}

func (d *Destination) Configure(ctx context.Context, cfg config.Config) (err error) {
	err = sdk.Util.ParseConfig(ctx, cfg, &d.config, d.config.Parameters())
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	d.config.logConfig(ctx, "configured destination")

	return nil
}

func (d *Destination) Open(ctx context.Context) (err error) {
	d.conn, err = connect(ctx, d.config.Config)
	if err != nil {
		return fmt.Errorf("failed to dial to ActiveMQ: %w", err)
	}
	sdk.Logger(ctx).Debug().Msg("opened destination")

	return nil
}

func (d *Destination) Write(ctx context.Context, records []opencdc.Record) (int, error) {
	for i, rec := range records {
		err := d.conn.Send(
			d.config.Destination, "application/json", rec.Bytes(),
			stomp.SendOpt.Receipt,
			stomp.SendOpt.Header("destination-type", d.config.DestinationType),
			stomp.SendOpt.Header("destination", d.config.DestinationHeader),
		)
		if err != nil {
			return i, fmt.Errorf("failed to send message: %w", err)
		}
		sdk.Logger(ctx).Trace().
			Str("destination", d.config.Destination).
			Str("destinationHeader", d.config.DestinationHeader).
			Str("destinationType", d.config.DestinationType).
			Msg("wrote record")
	}

	return len(records), nil
}

func (d *Destination) Teardown(ctx context.Context) error {
	return teardown(ctx, nil, d.conn, "destination")
}
