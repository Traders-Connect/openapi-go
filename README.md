# OpenAPI Go Client

A client library to work with cTrader's OpenAPI proxy servers.

## Install
```shell
go get github.com/ahmad-pepperstone/openapi-go
```


## Creating a client
```go
func OnMessage(data []byte) {
    message := &model.ProtoMessage{}
    err := proto.Unmarshal(data, message)
    if err != nil {
    fmt.Println(err)
      return
    }
    fmt.Println(message)
}

client := OpenAPI.NewClient(OpenAPI.ClientConfig{
    ClientID:     ClientID,
    ClientSecret: ClientSecret,
    Address:      Host,
    CertFile:     "server.crt",
    KeyFile:      "server.key",
})

```
### Connecting
```go
client.Connect()
```
### Disconnecting
```go
client.Disconnect()
```
### Subscribing to messages/events
```go
client.On("message", Handler)
```
### Unsubscribing
```go
client.Off("message", Handler)
```
### Sending messages
```go
client.SendMessage([]byte)
```

## Complete example with ProtoMessages
https://github.com/ahmad-pepperstone/ctrader-openapi-example
