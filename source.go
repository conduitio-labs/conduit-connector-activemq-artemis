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
	"errors"
	"fmt"
	"strings"

	"github.com/conduitio/conduit-commons/opencdc"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/go-stomp/stomp/v3"
	"github.com/go-stomp/stomp/v3/frame"
	"github.com/goccy/go-json"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type SourceConfig struct {
	Config

	sdk.DefaultSourceMiddleware

	// The size of the consumer window.
	// It maps to the "consumer-window-size" header in the STOMP SUBSCRIBE frame.
	ConsumerWindowSize string `json:"consumerWindowSize" default:"-1"`

	// The subscription type. It can be either ANYCAST or MULTICAST, with
	// ANYCAST being the default.
	// Maps to the "subscription-type" header in the STOMP SUBSCRIBE frame.
	SubscriptionType string `json:"subscriptionType" default:"ANYCAST" validation:"inclusion=ANYCAST|MULTICAST"`
}

type Source struct {
	sdk.UnimplementedSource
	config SourceConfig

	conn         *stomp.Conn
	subscription *stomp.Subscription

	storedMessages cmap.ConcurrentMap[string, *stomp.Message]
}

func (s *Source) Config() sdk.SourceConfig {
	return &s.config
}

func NewSource() sdk.Source {
	return sdk.SourceWithMiddleware(&Source{
		storedMessages: cmap.New[*stomp.Message](),
	})
}

func (s *Source) Open(ctx context.Context, sdkPos opencdc.Position) (err error) {
	s.conn, err = connect(ctx, s.config.Config)
	if err != nil {
		return fmt.Errorf("failed to dial to ActiveMQ: %w", err)
	}

	if sdkPos != nil {
		pos, err := parseSDKPosition(sdkPos)
		if err != nil {
			return fmt.Errorf("failed to parse position: %w", err)
		}

		if s.config.Destination != "" && s.config.Destination != pos.Destination {
			return fmt.Errorf(
				"the old position contains a different destination name than the connector configuration (%q vs %q), please check if the configured destination name changed since the last run",
				pos.Destination, s.config.Destination,
			)
		}

		sdk.Logger(ctx).Debug().Msg("got destination name from given position")
		s.config.Destination = pos.Destination
	}

	s.subscription, err = s.conn.Subscribe(s.config.Destination,
		stomp.AckClientIndividual,
		stomp.SubscribeOpt.Header("consumer-window-size", s.config.ConsumerWindowSize),
		stomp.SubscribeOpt.Header("subscription-type", s.config.SubscriptionType),
		stomp.SubscribeOpt.Header("destination", s.config.Destination),
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to destination: %w", err)
	}

	sdk.Logger(ctx).Debug().
		Str("destination", s.config.Destination).
		Str("subscriptionID", s.subscription.Id()).
		Msg("subscribed to destination")

	sdk.Logger(ctx).Debug().Msg("opened source")

	return nil
}

func (s *Source) Read(ctx context.Context) (opencdc.Record, error) {
	var rec opencdc.Record

	select {
	case <-ctx.Done():
		if err := ctx.Err(); err != nil {
			return rec, fmt.Errorf("context error: %w", err)
		}
		return rec, nil
	case msg, ok := <-s.subscription.C:
		if !ok {
			return rec, errors.New("source message channel closed")
		}

		if err := msg.Err; err != nil {
			return rec, fmt.Errorf("source message error: %w", err)
		}

		var (
			messageID = msg.Header.Get(frame.MessageId)
			pos       = Position{
				MessageID:   messageID,
				Destination: s.config.Destination,
			}
			sdkPos   = pos.ToSdkPosition()
			metadata = metadataFromMsg(msg)
			key      = opencdc.RawData(messageID)
			payload  = opencdc.RawData(msg.Body)
		)

		rec = sdk.Util.Source.NewRecordCreate(sdkPos, metadata, key, payload)

		sdk.Logger(ctx).Trace().
			Str("destination", s.config.Destination).
			Str("messageID", messageID).
			Str("destination", msg.Destination).
			Str("subscriptionDestination", msg.Subscription.Destination()).
			Msg("read message")

		s.storedMessages.Set(messageID, msg)

		return rec, nil
	}
}

// metadataFromMsg extracts all the present headers from a stomp.Message into
// opencdc.Metadata.
func metadataFromMsg(msg *stomp.Message) opencdc.Metadata {
	metadata := make(opencdc.Metadata)

	for i := range msg.Header.Len() {
		k, v := msg.Header.GetAt(i)

		// Prefix to avoid collisions with other metadata keys
		headerKey := "activemq.header." + k

		// According to the STOMP protocol, headers can have multiple values for
		// the same key. We concatenate them with a comma and a space.
		if headerVal, ok := metadata[headerKey]; ok {
			var sb strings.Builder
			sb.Grow(len(headerVal) + len(v) + 2)
			sb.WriteString(headerVal)
			sb.WriteString(", ")
			sb.WriteString(v)

			metadata[headerKey] = sb.String()
		} else {
			metadata[headerKey] = v
		}
	}

	return metadata
}

func (s *Source) Ack(ctx context.Context, position opencdc.Position) error {
	pos, err := parseSDKPosition(position)
	if err != nil {
		return fmt.Errorf("failed to parse position: %w", err)
	}

	msg, ok := s.storedMessages.Get(pos.MessageID)
	if !ok {
		return fmt.Errorf("message with ID %q not found", pos.MessageID)
	}

	if err := s.conn.Ack(msg); err != nil {
		return fmt.Errorf("failed to ack message: %w", err)
	}

	s.storedMessages.Pop(pos.MessageID)

	sdk.Logger(ctx).Trace().Str("destination", s.config.Destination).Msgf("acked message")

	return nil
}

func (s *Source) Teardown(ctx context.Context) error {
	return teardown(ctx, s.subscription, s.conn, "source")
}

type Position struct {
	MessageID   string `json:"message_id"`
	Destination string `json:"destination"`
}

func parseSDKPosition(sdkPos opencdc.Position) (Position, error) {
	decoder := json.NewDecoder(bytes.NewBuffer(sdkPos))
	decoder.DisallowUnknownFields()

	var p Position
	if err := decoder.Decode(&p); err != nil {
		return p, fmt.Errorf("failed to decode position: %w", err)
	}
	return p, nil
}

func (p Position) ToSdkPosition() opencdc.Position {
	bs, err := json.Marshal(p)
	if err != nil {
		// this should never happen
		panic(err)
	}

	return opencdc.Position(bs)
}
