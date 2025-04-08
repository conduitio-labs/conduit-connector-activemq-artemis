# Conduit Connector for Activemq Artemis

<!-- readmegen:description -->
The ActiveMQ Classic connector is one of [Conduit](https://conduit.io) plugins. The connector provides both a source and a destination connector for [ActiveMQ Artemis](https://activemq.apache.org/components/artemis/).

It uses the [stomp protocol](https://stomp.github.io/) to connect to ActiveMQ.

## What data does the OpenCDC record consist of?

- `record.Position`: json object with the destination and the messageId frame header.
- `record.Operation`: currently fixed as "create".
- `record.Metadata`: a string to string map, with keys prefixed as `activemq.header.{STOMP_HEADER_NAME}`.
- `record.Key`: the messageId frame header.
- `record.Payload.Before`: empty
- `record.Payload.After`: the message body<!-- /readmegen:description -->

## How to build?

Run `make build` to build the connector.

## Testing

Run `make test` to run all the tests. The command will handle starting and stopping docker containers for you.

## Source configuration

<!-- readmegen:source.parameters.yaml -->
```yaml
version: 2.2
pipelines:
  - id: example
    status: running
    connectors:
      - id: example
        plugin: "activemq"
        settings:
          # Destination is the name of the STOMP destination.
          # Type: string
          # Required: yes
          destination: ""
          # Password is the password to use when connecting to the broker.
          # Type: string
          # Required: yes
          password: ""
          # URL is the URL of the ActiveMQ Artemis broker.
          # Type: string
          # Required: yes
          url: ""
          # User is the username to use when connecting to the broker.
          # Type: string
          # Required: yes
          user: ""
          # ConsumerWindowSize is the size of the consumer window. It maps to
          # the "consumer-window-size" header in the STOMP SUBSCRIBE frame.
          # Type: string
          # Required: no
          consumerWindowSize: "-1"
          # RecvTimeoutHeartbeat specifies the minimum amount of time between
          # the client expecting to receive heartbeat notifications from the
          # server
          # Type: duration
          # Required: no
          recvTimeoutHeartbeat: "2s"
          # SendTimeoutHeartbeat specifies the maximum amount of time between
          # the client sending heartbeat notifications from the server
          # Type: duration
          # Required: no
          sendTimeoutHeartbeat: "2s"
          # SubscriptionType is the subscription type. It can be either ANYCAST
          # or MULTICAST, with ANYCAST being the default. Maps to the
          # "subscription-type" header in the STOMP SUBSCRIBE frame.
          # Type: string
          # Required: no
          subscriptionType: "ANYCAST"
          # CaCertPath is the path to the CA certificate file.
          # Type: string
          # Required: no
          tls.caCertPath: ""
          # ClientCertPath is the path to the client certificate file.
          # Type: string
          # Required: no
          tls.clientCertPath: ""
          # ClientKeyPath is the path to the client key file.
          # Type: string
          # Required: no
          tls.clientKeyPath: ""
          # Enabled is a flag to enable or disable TLS.
          # Type: bool
          # Required: no
          tls.enabled: "false"
          # InsecureSkipVerify is a flag to disable server certificate
          # verification.
          # Type: bool
          # Required: no
          tls.insecureSkipVerify: "false"
          # Maximum delay before an incomplete batch is read from the source.
          # Type: duration
          # Required: no
          sdk.batch.delay: "0"
          # Maximum size of batch before it gets read from the source.
          # Type: int
          # Required: no
          sdk.batch.size: "0"
          # Specifies whether to use a schema context name. If set to false, no
          # schema context name will be used, and schemas will be saved with the
          # subject name specified in the connector (not safe because of name
          # conflicts).
          # Type: bool
          # Required: no
          sdk.schema.context.enabled: "true"
          # Schema context name to be used. Used as a prefix for all schema
          # subject names. If empty, defaults to the connector ID.
          # Type: string
          # Required: no
          sdk.schema.context.name: ""
          # Whether to extract and encode the record key with a schema.
          # Type: bool
          # Required: no
          sdk.schema.extract.key.enabled: "true"
          # The subject of the key schema. If the record metadata contains the
          # field "opencdc.collection" it is prepended to the subject name and
          # separated with a dot.
          # Type: string
          # Required: no
          sdk.schema.extract.key.subject: "key"
          # Whether to extract and encode the record payload with a schema.
          # Type: bool
          # Required: no
          sdk.schema.extract.payload.enabled: "true"
          # The subject of the payload schema. If the record metadata contains
          # the field "opencdc.collection" it is prepended to the subject name
          # and separated with a dot.
          # Type: string
          # Required: no
          sdk.schema.extract.payload.subject: "payload"
          # The type of the payload schema.
          # Type: string
          # Required: no
          sdk.schema.extract.type: "avro"
```
<!-- /readmegen:source.parameters.yaml -->

## Destination configuration

<!-- readmegen:destination.parameters.yaml -->
```yaml
version: 2.2
pipelines:
  - id: example
    status: running
    connectors:
      - id: example
        plugin: "activemq"
        settings:
          # Destination is the name of the STOMP destination.
          # Type: string
          # Required: yes
          destination: ""
          # Password is the password to use when connecting to the broker.
          # Type: string
          # Required: yes
          password: ""
          # URL is the URL of the ActiveMQ Artemis broker.
          # Type: string
          # Required: yes
          url: ""
          # User is the username to use when connecting to the broker.
          # Type: string
          # Required: yes
          user: ""
          # DestinationHeader maps to the "destination" header in the STOMP SEND
          # frame. Useful when using ANYCAST.
          # Type: string
          # Required: no
          destinationHeader: ""
          # DestinationType is the routing type of the destination. It can be
          # either ANYCAST or MULTICAST, with ANYCAST being the default. Maps to
          # the "destination-type" header in the STOMP SEND frame.
          # Type: string
          # Required: no
          destinationType: "ANYCAST"
          # RecvTimeoutHeartbeat specifies the minimum amount of time between
          # the client expecting to receive heartbeat notifications from the
          # server
          # Type: duration
          # Required: no
          recvTimeoutHeartbeat: "2s"
          # SendTimeoutHeartbeat specifies the maximum amount of time between
          # the client sending heartbeat notifications from the server
          # Type: duration
          # Required: no
          sendTimeoutHeartbeat: "2s"
          # CaCertPath is the path to the CA certificate file.
          # Type: string
          # Required: no
          tls.caCertPath: ""
          # ClientCertPath is the path to the client certificate file.
          # Type: string
          # Required: no
          tls.clientCertPath: ""
          # ClientKeyPath is the path to the client key file.
          # Type: string
          # Required: no
          tls.clientKeyPath: ""
          # Enabled is a flag to enable or disable TLS.
          # Type: bool
          # Required: no
          tls.enabled: "false"
          # InsecureSkipVerify is a flag to disable server certificate
          # verification.
          # Type: bool
          # Required: no
          tls.insecureSkipVerify: "false"
          # Maximum delay before an incomplete batch is written to the
          # destination.
          # Type: duration
          # Required: no
          sdk.batch.delay: "0"
          # Maximum size of batch before it gets written to the destination.
          # Type: int
          # Required: no
          sdk.batch.size: "0"
          # Allow bursts of at most X records (0 or less means that bursts are
          # not limited). Only takes effect if a rate limit per second is set.
          # Note that if `sdk.batch.size` is bigger than `sdk.rate.burst`, the
          # effective batch size will be equal to `sdk.rate.burst`.
          # Type: int
          # Required: no
          sdk.rate.burst: "0"
          # Maximum number of records written per second (0 means no rate
          # limit).
          # Type: float
          # Required: no
          sdk.rate.perSecond: "0"
          # The format of the output record. See the Conduit documentation for a
          # full list of supported formats
          # (https://conduit.io/docs/using/connectors/configuration-parameters/output-format).
          # Type: string
          # Required: no
          sdk.record.format: "opencdc/json"
          # Options to configure the chosen output record format. Options are
          # normally key=value pairs separated with comma (e.g.
          # opt1=val2,opt2=val2), except for the `template` record format, where
          # options are a Go template.
          # Type: string
          # Required: no
          sdk.record.format.options: ""
          # Whether to extract and decode the record key with a schema.
          # Type: bool
          # Required: no
          sdk.schema.extract.key.enabled: "true"
          # Whether to extract and decode the record payload with a schema.
          # Type: bool
          # Required: no
          sdk.schema.extract.payload.enabled: "true"
```
<!-- /readmegen:destination.parameters.yaml -->
