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
	"testing"

	"github.com/go-stomp/stomp/v3"
	"github.com/go-stomp/stomp/v3/frame"
	"github.com/matryer/is"
)

func TestMetadataFromMsg(t *testing.T) {
	is := is.New(t)

	header := &frame.Header{}
	header.Add("key", "value1")
	header.Add("key", "value2")

	header.Add("key2", "value3")
	header.Add("key3", "value4")

	metadata := metadataFromMsg(&stomp.Message{
		Header: header,
	})

	is.Equal(metadata["activemq.header.key"], "value1, value2")
	is.Equal(metadata["activemq.header.key2"], "value3")
	is.Equal(metadata["activemq.header.key3"], "value4")
}
