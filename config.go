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
	"time"
)

type Config struct {
	// URL is the URL of the ActiveMQ Artemis broker.
	URL string `json:"url" validate:"required"`

	// User is the username to use when connecting to the broker.
	User string `json:"user" validate:"required"`

	// Password is the password to use when connecting to the broker.
	Password string `json:"password" validate:"required"`

	// Destination is the name of the STOMP destination.
	Destination string `json:"destination" validate:"required"`

	// SendTimeoutHeartbeat specifies the maximum amount of time between the
	// client sending heartbeat notifications from the server
	SendTimeoutHeartbeat time.Duration `json:"sendTimeoutHeartbeat" default:"2s"`

	// RecvTimeoutHeartbeat specifies the minimum amount of time between the
	// client expecting to receive heartbeat notifications from the server
	RecvTimeoutHeartbeat time.Duration `json:"recvTimeoutHeartbeat" default:"2s"`

	TLS TLSConfig `json:"tls"`
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
