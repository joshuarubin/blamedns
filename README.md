# blamedns

## TODO

1. api server
1. websocket server
1. web interface
1. wildcard matching of (sub)domains for block and whitelist
1. ensure no duplicates of resp.Answer are in resp.Extra
1. resolve recursive requests using root.hints (if no forwards)
1. dnssec validation
    * if validated, set resp.AuthenticatedData
1. negative response caching (based on SOA TTL)
1. add zone parsing and authoritative responses for them
1. pattern matching for blocked domains
