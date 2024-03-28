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
	"math/rand"
	"testing"

	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

// cfgToMap converts a config struct to a map. This is useful for more type
// safety on tests.
func cfgToMap(cfg any) map[string]string {
	bs, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	mAny := map[string]any{}
	err = json.Unmarshal(bs, &mAny)
	if err != nil {
		panic(err)
	}

	m := map[string]string{}
	for k, v := range mAny {
		switch v := v.(type) {
		case string:
			m[k] = v
		case bool:
			m[k] = fmt.Sprintf("%t", v)
		case float64:
			// using %v to avoid scientific notation
			m[k] = fmt.Sprintf("%v", v)
		case map[string]any:
			parsed := cfgToMap(v)
			for k2, v := range parsed {
				m[k+"."+k2] = v
			}
		default:
			panic(fmt.Errorf("unsupported type used for cfgToMap func: %T", v))
		}
	}

	return m
}

func generateSDKRecords(from, to int) []sdk.Record {
	var sdkRecs []sdk.Record
	for i := from; i < to; i++ {
		key := []byte(fmt.Sprintf("test-key-%d", i))
		value := []byte(fmt.Sprintf("test-payload-%d", i))

		sdkRecs = append(sdkRecs, sdk.Util.Source.NewRecordCreate(
			[]byte(uuid.NewString()),
			sdk.Metadata{},
			sdk.RawData(key),
			sdk.RawData(value),
		))
	}
	return sdkRecs
}

func produce(is *is.I, cfgMap map[string]string, recs []sdk.Record) {
	ctx := context.Background()

	destination := NewDestination()
	err := destination.Configure(ctx, cfgMap)

	err = destination.Open(ctx)
	is.NoErr(err)

	totalWritten, err := destination.Write(ctx, recs)
	is.Equal(totalWritten, len(recs))

	is.NoErr(err)

	err = destination.Teardown(ctx)
	is.NoErr(err)
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func randBytes() []byte {
	return []byte(randString())
}

func uniqueDestinationName(t *testing.T) string {
	return fmt.Sprintf("/queue/%s_%s", t.Name(), uuid.New().String()[0:8])
}
