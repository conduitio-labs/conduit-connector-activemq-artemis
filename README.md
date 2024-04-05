# Conduit Connector for Activemq Artemis

The ActiveMQ Classic connector is one of [Conduit](https://conduit.io) plugins. The connector provides both a source and a destination connector for [ActiveMQ Artemis](https://activemq.apache.org/components/artemis/).

It uses the [stomp protocol](https://stomp.github.io/) to connect to ActiveMQ.

## What data does the OpenCDC record consist of?

| Field                   | Description                                                                          |
|-------------------------|--------------------------------------------------------------------------------------|
| `record.Position`       | json object with the destination and the messageId frame header.                     |
| `record.Operation`      | currently fixed as "create".                                                         |
| `record.Metadata`       | a string to string map, with keys prefixed as `activemq.header.{STOMP_HEADER_NAME}`. |
| `record.Key`            | the messageId frame header.                                                          |
| `record.Payload.Before` | <empty>                                                                              |
| `record.Payload.After`  | the message body                                                                     |


## How to build?
Run `make build` to build the connector.

## Testing
Run `make test` to run all the tests. The command will handle starting and stopping docker containers for you.

## Configuration

Both source and destination connectors share the following parameters:

| name | description | required | default value |
| ---- | ----------- | -------- | ------------- |
| `url` | URL of the ActiveMQ Artemis broker. | true |  |
| `user` | The username to use when connecting to the broker. | true |  |
| `password` | The password to use when connecting to the broker. | true |  |
| `destination` | The name of the STOMP destination. | true |  |
| `sendTimeoutHeartbeat` | Specifies the maximum amount of time between the client sending heartbeat notifications from the server. | true | 2s (*) |
| `recvTimeoutHeartbeat` | Specifies the minimum amount of time between the client expecting to receive heartbeat notifications from the server. | true | 2s (*) |
| `tls.enabled` | Flag to enable or disable TLS. | false | `false` |
| `tls.clientKeyPath` | Path to the client key file. | false |  |
| `tls.clientCertPath` | Path to the client certificate file. | false |  |
| `tls.caCertPath` | Path to the CA certificate file. | false |  |
| `tls.insecureSkipVerify` | Flag to skip verification of the server's certificate chain and host name | false |  |

(*) Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".


### Source configuration

| name | description | required | default value |
| ---- | ----------- | -------- | ------------- |
| `consumerWindowSize` | The size of the consumer window. It maps to the "consumer-window-size" header in the STOMP SUBSCRIBE frame. | false | -1 |
| `subscriptionType` | The subscription type. It can be either ANYCAST or MULTICAST, with ANYCAST being the default. Maps to the "subscription-type" header in the STOMP SUBSCRIBE frame. | false | ANYCAST |

### Destination configuration

| name | description | required | default value |
| ---- | ----------- | -------- | ------------- |
| `destinationType` | The routing type of the destination. It can be either ANYCAST or MULTICAST, with ANYCAST being the default. Maps to the "destination-type" header in the STOMP SEND frame. | false | ANYCAST |
| `destinationHeader` | Maps to the "destination" header in the STOMP SEND frame. Useful when using ANYCAST. | false | |


## Example pipeline.yml file

Here's an example of a `pipeline.yml` file using `file to activemq artemis` and `activemq artemis to file` pipelines: 

```yaml
version: 2.0
pipelines:
  - id: file-to-activemq-artemis
    status: running
    connectors:
      - id: file.in
        type: source
        plugin: builtin:file
        name: file-destination
        settings:
          path: ./file.in
      - id: activemq-artemis.out
        type: destination
        plugin: standalone:activemq@latest
        name: activemq-artemis-source
        settings:
          url: localhost:61613
          user: admin
          password: admin
          destination: example
          destinationType: ANYCAST

          sdk.record.format: template
          sdk.record.format.options: '{{ printf "%s" .Payload.After }}'

  - id: activemq-artemis-to-file
    status: running
    connectors:
      - id: activemq-artemis.in
        type: source
        plugin: standalone:activemq@latest
        name: activemq-artemis-source
        settings:
          url: localhost:61613
          user: admin
          password: admin
          destination: example
          subscriptionType: ANYCAST
          consumerWindowSize: -1

      - id: file.out
        type: destination
        plugin: builtin:file
        name: file-destination
        settings:
          path: ./file.out
          sdk.record.format: template
          sdk.record.format.options: '{{ printf "%s" .Payload.After }}'
```
