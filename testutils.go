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
	"encoding/json"
	"fmt"
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
