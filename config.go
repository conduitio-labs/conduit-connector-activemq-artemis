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
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

//go:generate paramgen -output=paramgen_src.go SourceConfig
//go:generate paramgen -output=paramgen_dest.go DestinationConfig

type Config struct {
	// URL is the URL of the ActiveMQ classic broker.
	URL string `json:"url" validate:"required"`

	// User is the username to use when connecting to the broker.
	User string `json:"user" validate:"required"`

	// Password is the password to use when connecting to the broker.
	Password string `json:"password" validate:"required"`

	// Destination is the name of the destination to read from.
	Destination string `json:"destination" validate:"required"`

	// SendTimeoutHeartbeat specifies the maximum amount of time between the
	// client sending heartbeat notifications from the server
	SendTimeoutHeartbeat time.Duration `json:"sendTimeoutHeartbeat" default:"2s"`

	// RecvTimeoutHeartbeat specifies the minimum amount of time between the
	// client expecting to receive heartbeat notifications from the server
	RecvTimeoutHeartbeat time.Duration `json:"recvTimeoutHeartbeat" default:"2s"`

	TLS TLSConfig `json:"tls"`
}

func (c Config) logConfig(ctx context.Context, msg string) {
	sdk.Logger(ctx).Debug().
		Str("url", c.URL).
		Str("destination", c.Destination).
		Str("sendTimeoutHeartbeat", c.SendTimeoutHeartbeat.String()).
		Str("recvTimeoutHeartbeat", c.RecvTimeoutHeartbeat.String()).
		Bool("tlsEnabled", c.TLS.Enabled).Msg(msg)
}

type TLSConfig struct {
	// Enabled is a flag to enable or disable TLS.
	Enabled bool `json:"enabled" default:"false"`

	// ClientKeyPath is the path to the client key file.
	ClientKeyPath string `json:"clientKeyPath"`

	// ClientCertPath is the path to the client certificate file.
	ClientCertPath string `json:"clientCertPath"`

	// CaCertPath is the path to the CA certificate file.
	CaCertPath string `json:"caCertPath"`

	// InsecureSkipVerify is a flag to disable server certificate verification.
	InsecureSkipVerify bool `json:"insecureSkipVerify" default:"false"`
}

type SourceConfig struct {
	Config

	// ConsumerWindowSize is the size of the consumer window.
	// It maps to the "consumer-window-size" header in the STOMP SUBSCRIBE frame.
	ConsumerWindowSize string `json:"consumerWindowSize" default:"-1"`

	// SubscriptionType is the subscription type. It can be either
	// ANYCAST or MULTICAST, with ANYCAST being the default.
	// Maps to the "subscription-type" header in the STOMP SUBSCRIBE frame.
	SubscriptionType string `json:"subscriptionType" default:"ANYCAST" validation:"inclusion=ANYCAST|MULTICAST"`
}

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
