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
	"time"

	sdk "github.com/conduitio/conduit-connector-sdk"
)

func TestAcceptanceANYCAST(t *testing.T) {
	sourceConfig := map[string]string{
		"url":                "localhost:61613",
		"user":               "admin",
		"password":           "admin",
		"consumerWindowSize": "-1",
		"subscriptionType":   "ANYCAST",
	}
	destinationConfig := map[string]string{
		"url":             "localhost:61613",
		"user":            "admin",
		"password":        "admin",
		"destinationType": "ANYCAST",
	}

	driver := sdk.ConfigurableAcceptanceTestDriver{
		Config: sdk.ConfigurableAcceptanceTestDriverConfig{
			Connector:         Connector,
			SourceConfig:      sourceConfig,
			DestinationConfig: destinationConfig,
			BeforeTest: func(t *testing.T) {
				destination := uniqueDestinationName(t)
				sourceConfig["destination"] = destination
				destinationConfig["destination"] = destination
				destinationConfig["destinationHeader"] = destination
			},
			Skip: []string{
				"TestSource_Configure_RequiredParams",
				"TestDestination_Configure_RequiredParams",
			},
			WriteTimeout: 500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
		},
	}

	sdk.AcceptanceTest(t, driver)
}

func TestAcceptanceMULTICAST(t *testing.T) {
	sourceConfig := map[string]string{
		"url":                "localhost:61613",
		"user":               "admin",
		"password":           "admin",
		"consumerWindowSize": "-1",
		"subscriptionType":   "MULTICAST",
	}
	destinationConfig := map[string]string{
		"url":             "localhost:61613",
		"user":            "admin",
		"password":        "admin",
		"destinationType": "MULTICAST",
	}

	driver := sdk.ConfigurableAcceptanceTestDriver{
		Config: sdk.ConfigurableAcceptanceTestDriverConfig{
			Connector:         Connector,
			SourceConfig:      sourceConfig,
			DestinationConfig: destinationConfig,
			BeforeTest: func(t *testing.T) {
				destination := uniqueDestinationName(t)
				sourceConfig["destination"] = destination
				destinationConfig["destination"] = destination
			},
			Skip: []string{
				"TestSource_Configure_RequiredParams",
				"TestDestination_Configure_RequiredParams",
			},
			WriteTimeout: 500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
		},
	}

	sdk.AcceptanceTest(t, driver)
}
