# eventbus
Straightforward EventBus for Go 1.18+

[![Documentation](https://pkg.go.dev/badge/github.com/JosiahWitt/eventbus)](https://pkg.go.dev/github.com/JosiahWitt/eventbus)
[![CI](https://github.com/JosiahWitt/eventbus/workflows/CI/badge.svg)](https://github.com/JosiahWitt/eventbus/actions?query=branch%3Amain+workflow%3ACI)
[![codecov](https://codecov.io/gh/JosiahWitt/eventbus/branch/main/graph/badge.svg)](https://codecov.io/gh/JosiahWitt/eventbus)
<!-- Hiding the Go Report Card until they support generics -->
<!-- [![Go Report Card](https://goreportcard.com/badge/github.com/JosiahWitt/eventbus)](https://goreportcard.com/report/github.com/JosiahWitt/eventbus) -->

## Go Version Support
Go 1.18+ is required, due to the use of generics


## Project Status
Unlike some of my other projects that are used in production (eg. [ensure](https://github.com/JosiahWitt/ensure) for testing and [erk](https://github.com/JosiahWitt/erk) for errors), this project is _currently more on the experimental side_. It has good test coverage, but I haven't actually built anything serious with it (yet!). So, please **use at your own risk**! Oh, and if you do use it for something serious, let me know how it works out!


## Install
### Library
```bash
$ go get github.com/JosiahWitt/eventbus
```


## About
This package provides a straightforward concurrent, typed EventBus for Go 1.18+, supporting fanout, and in order, only once delivery.


## Examples

Please see the [`examples`](./examples/) directory for complete examples, such as the [Simple Chat App](./examples/simplechatapp/).

### Basic Usage
```go
type Message struct { Body string }

bus := eventbus.New[*Message]()

// In one goroutine:
sub := bus.Subscribe("chatroom:123", "user:456")
for msg := range sub.Channel() {
  // Do something with msg
}
// Eventually...
sub.Unsubscribe()

// In another goroutine:
bus.Publish(&Message{Body: "Hello, World!"}, "chatroom:123", "user:789")
bus.Publish(&Message{Body: "Go is fun!"}, "chatroom:456", "user:456")
bus.Publish(&Message{Body: "Wordle 1/6"}, "chatroom:456", "user:789")
bus.Publish(&Message{Body: "Bonjour!"}, "chatroom:123", "user:456")

// sub will receive:
//  1. Hello, World!
//  2. Go is fun!
//  3. Bonjour!
// Each of those messages is published to one or both of `chatroom:123` and `user:456`.
//
// sub will not receive "Wordle 1/6", because that was published to neither `chatroom:123` nor `user:456`.
```
