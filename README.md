# blamedns

[![CircleCI](https://circleci.com/gh/joshuarubin/blamedns.svg?style=svg)](https://circleci.com/gh/joshuarubin/blamedns) [![Coverage Status](https://coveralls.io/repos/github/joshuarubin/blamedns/badge.svg?branch=master)](https://coveralls.io/github/joshuarubin/blamedns?branch=master) [![Go Report Card](https://goreportcard.com/badge/jrubin.io/blamedns)](https://goreportcard.com/report/jrubin.io/blamedns) [![GoDoc](https://godoc.org/jrubin.io/blamedns?status.svg)](https://godoc.org/jrubin.io/blamedns) ![License](https://img.shields.io/badge/license-apache-blue.svg)

## TODO

1. defined exit codes
1. shell autocomplete
1. stop using defer to unlock mutexes
1. verify dl files are updated on interval
1. api server
    * clear cache
    * add to whitelist
1. web interface
1. testing
1. documentation
1. profiling
1. fuzzing
1. blamedns.com website

```sh
sudo setcap cap_net_bind_service=+ep /usr/local/bin/blamedns
sudo adduser --system blamedns --home /var/cache/blamedns --shell /usr/sbin/nologin --disabled-password --disabled-login --group --no-create-home
```
