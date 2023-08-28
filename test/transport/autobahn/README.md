# Autobahn Server

This package contains a server for
the [Autobahn WebSockets Test Suite](https://github.com/crossbario/autobahn-testsuite).

To test the server, run

```bash
go run . [--addr=<addr>]
```

> where `<addr>` is the address to listen on. The default is `:9000`.
> > **Note** that if you are using custom address you
> > should also update the test suite config [here](config/fuzzingclient.json) putting actual address of your server there.

and start the client test driver

```bash
mkdir -p reports
docker run -it --rm \
  -v ${PWD}/config:/config \
  -v ${PWD}/reports:/reports \
  crossbario/autobahn-testsuite \
  wstest -m fuzzingclient -s /config/fuzzingclient.json

```

When the client completes, it writes a report to `reports/index.html`.