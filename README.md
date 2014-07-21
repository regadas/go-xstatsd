## go-xstatsd ##

This is a very simple/small statsd client that can also allow sending stats in bulk.

**Disclaimer**

This client is based upon the [example](https://github.com/etsy/statsd/tree/master/examples/go) provided by the etsy folks. Thank you etsy!


### Simple example ###
```go
import "github.com/regadas/go-xstatsd"

s := statsd.New("127.0.0.1:8125", "some.metric.prefix")

s.Timing("foobar", 16)
```

### Bulk example ###
```go
import (
    "net"
    "github.com/regadas/go-xstatsd"
)

s := statsd.New("127.0.0.1:8125", "some.metric.prefix")

s.Client.WithConnection(func(conn *net.Conn) {
	s.TimingRaw(conn, "foobar", 16)
	s.TimingRaw(conn, "barfoo", 32)
	s.IncrementRaw(conn, "bar")
})
```

