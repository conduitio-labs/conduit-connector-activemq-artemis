package main

import (
	sdk "github.com/conduitio/conduit-connector-sdk"

	activemq "github.com/alarbada/conduit-connector-activemq-artemis"
)

func main() {
	sdk.Serve(activemq.Connector)
}
