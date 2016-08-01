<img src="https://rawgit.com/mvader/flamingo/master/flamingo.png" width="400" height="200" />

[![godoc reference](https://godoc.org/github.com/mvader/flamingo?status.png)](https://godoc.org/github.com/mvader/flamingo)
Flamingo is a very thin and simple bot framework layer on top of [nlopes/slack](https://github.com/nlopes/slack) Slack API.

Even though [nlopes/slack](https://github.com/nlopes/slack) is hands down probably the best Slack client library out there for Golang, it's a bit tedious to manage a bot with several commands using the RTM connection on your own, not to mention operations as simple as replying to an user in a channel.
Flamingo aims to provide an API as thin as possible to build a bot on top of [nlopes/slack](https://github.com/nlopes/slack) and make as easy as possible without adding much overhead.

You can see a simple example at [examples folder](https://github.com/mvader/flamingo/blob/master/examples/hello.go).

## TODO

* Unit/integration tests
