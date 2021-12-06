[![Build Status](https://github.com/go-pkgz/auth/workflows/build/badge.svg)](https://github.com/butwhoare-you/rynek-pierwotny-updates-cli/actions)
[![Coverage Status](https://coveralls.io/repos/github/butwhoareyou/rynek-pierwotny-updates-cli/badge.svg?branch=master)](https://coveralls.io/github/butwhoareyou/rynek-pierwotny-updates-cli?branch=master)

# RynekPierwotny Updates CLI

## Commands

### offers-updates

* Usage without docker

```shell
rynek-pierwotny-updates-cli offers-updates \
--aws.region="eu-west-1" \
--aws.s3.bucket="offer-updates-1" \
--aws.endpoint="http://localhost:9000" \
--request.regions=120 \
--url="https://rynekpierwotny.pl" \
--api-url="https://rynekpierwotny.pl/api" \
--debug
```

This command fetches all offers for provided regions.
It stores fetched offers in local S3-like storage (min.io)
