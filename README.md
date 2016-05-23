# blamedns

## TODO

1. app.Usage
1. reorganize config opts
    * see unbound config
1. api server
1. import _ "net/http/pprof"
1. profiling
1. fuzzing
1. websocket server
1. web interface
1. resolve recursive requests using root.hints (if no forwards)
1. dnssec validation
    * if validated, set resp.AuthenticatedData
1. negative response caching (based on SOA TTL)
1. add zone parsing and authoritative responses for them
