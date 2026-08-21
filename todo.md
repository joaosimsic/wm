- [ ] too many _ with no handled errors, this misses the point of returning errors
- [ ] manager is doing too much
- [ ] implement constructor functions to enforce comptime errors if struct misuse:

```go
type Config struct {
  Host string
  Port int
  TLS  bool
}

func NewConfig(host string, port int, tls bool) Config {
    return Config{
        Host: host,
        Port: port,
        TLS: tls,
    }
}

```
```
