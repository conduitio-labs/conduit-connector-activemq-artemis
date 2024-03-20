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
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/go-stomp/stomp/v3"
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

	// Queue is the name of the queue to read from.
	Queue string `json:"queue" validate:"required"`

	// ContentType is the content type of the message.
	ContentType string `json:"contentType" default:"text/plain"`

	// SendTimeoutHeartbeat specifies the maximum amount of time between the
	// client sending heartbeat notifications from the server
	SendTimeoutHeartbeat time.Duration `json:"sendTimeoutHeartbeat" default:"2s"`

	// RecvTimeoutHeartbeat specifies the minimum amount of time between the
	// client expecting to receive heartbeat notifications from the server
	RecvTimeoutHeartbeat time.Duration `json:"recvTimeoutHeartbeat" default:"2s"`

	TLS TLSConfig `json:"tlsConfig"`
}

type TLSConfig struct {
	// UseTLS is a flag to enable or disable TLS.
	UseTLS bool `json:"useTLS" default:"false"`

	// ClientKeyPath is the path to the client key file.
	ClientKeyPath string `json:"clientKeyPath"`

	// ClientCertPath is the path to the client certificate file.
	ClientCertPath string `json:"clientCertPath"`

	// CaCertPath is the path to the CA certificate file.
	CaCertPath string `json:"caCertPath"`
}

type SourceConfig struct {
	Config
}

type DestinationConfig struct {
	Config
}

type Position struct {
	MessageID string `json:"message_id"`
	Queue     string `json:"queue"`
}

func parseSDKPosition(sdkPos sdk.Position) (Position, error) {
	decoder := json.NewDecoder(bytes.NewBuffer(sdkPos))
	decoder.DisallowUnknownFields()

	var p Position
	err := decoder.Decode(&p)
	return p, err
}

func (p Position) ToSdkPosition() sdk.Position {
	bs, err := json.Marshal(p)
	if err != nil {
		// this should never happen
		panic(err)
	}

	return sdk.Position(bs)
}

// metadataFromMsg extracts all the present headers from a stomp.Message into
// sdk.Metadata.
func metadataFromMsg(msg *stomp.Message) sdk.Metadata {
	metadata := make(sdk.Metadata)

	for i := range msg.Header.Len() {
		k, v := msg.Header.GetAt(i)
		metadata[k] = v
	}

	return metadata
}

func connect(ctx context.Context, config Config) (*stomp.Conn, error) {
	connOpts := []func(*stomp.Conn) error{
		stomp.ConnOpt.Login(config.User, config.Password),
		stomp.ConnOpt.HeartBeat(config.SendTimeoutHeartbeat, config.RecvTimeoutHeartbeat),
	}

	if !config.TLS.UseTLS {
		conn, err := stomp.Dial("tcp", config.URL, connOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to ActiveMQ: %w", err)
		}
		sdk.Logger(ctx).Debug().Msg("opened connection to ActiveMQ")

		return conn, nil
	}

	sdk.Logger(ctx).Debug().Msg("using TLS to connect to ActiveMQ")

	cert, err := tls.LoadX509KeyPair(config.TLS.ClientCertPath, config.TLS.ClientKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client key pair: %w", err)
	}
	sdk.Logger(ctx).Debug().Msg("loaded client key pair")

	caCert, err := os.ReadFile(config.TLS.CaCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA cert: %w", err)
	}
	sdk.Logger(ctx).Debug().Msg("loaded CA cert")

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,

		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	// version will be overwritten at compile time when building a release,
	// so this should only be true when running in development mode.
	if version == "(devel)" {
		tlsConfig.InsecureSkipVerify = true
	}

	netConn, err := tls.Dial("tcp", config.URL, tlsConfig)
	if err != nil {
		panic(err)
	}
	sdk.Logger(ctx).Debug().Msg("TLS connection established")

	conn, err := stomp.Connect(netConn, connOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ActiveMQ: %w", err)
	}
	sdk.Logger(ctx).Debug().Msg("STOMP connection using tls established")

	return conn, nil
}

func teardown(ctx context.Context, subs *stomp.Subscription, conn *stomp.Conn, label string) error {
	if subs != nil {
		err := subs.Unsubscribe()
		if errors.Is(err, stomp.ErrCompletedSubscription) {
			sdk.Logger(ctx).Debug().Msg("subscription already unsubscribed")
		} else if err != nil {
			return fmt.Errorf("failed to unsubscribe: %w", err)
		}
	}

	if conn != nil {
		if err := conn.Disconnect(); err != nil {
			return fmt.Errorf("failed to disconnect from ActiveMQ: %w", err)
		}
	}

	sdk.Logger(ctx).Debug().Msgf("teardown for %s complete", label)

	return nil
}
