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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/go-stomp/stomp/v3"
)

func connect(ctx context.Context, config Config) (*stomp.Conn, error) {
	connOpts := []func(*stomp.Conn) error{
		stomp.ConnOpt.Login(config.User, config.Password),
		stomp.ConnOpt.HeartBeat(config.SendTimeoutHeartbeat, config.RecvTimeoutHeartbeat),
	}

	if !config.TLS.Enabled {
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

		InsecureSkipVerify: config.TLS.InsecureSkipVerify, // #nosec G402
	}

	netConn, err := tls.Dial("tcp", config.URL, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ActiveMQ using tls: %w", err)
	}
	sdk.Logger(ctx).Debug().Msg("TLS connection established")

	conn, err := stomp.Connect(netConn, connOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ActiveMQ with tls connection: %w", err)
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
