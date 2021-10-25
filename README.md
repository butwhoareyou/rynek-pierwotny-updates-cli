[![Build Status](https://github.com/go-pkgz/auth/workflows/build/badge.svg)](https://github.com/butwhoare-you/rynek-pierwotny-updates-cli/actions)
[![Coverage Status](https://coveralls.io/repos/github/butwhoareyou/rynek-pierwotny-updates-cli/badge.svg?branch=master)](https://coveralls.io/github/butwhoareyou/rynek-pierwotny-updates-cli?branch=master)

# RynekPierwotny Updates CLI

## Commands

### offers-updates

* Usage without docker

```shell
rynek-pierwotny-updates-cli offers-updates \
--request.regions=1 \
--url=https://rynekpierwotny.pl \
--api-url=https://rynekpierwotny.pl/api \
--telegram-chat-id=123 \ 
--telegram-token="TOKEN" \
--fs-store-path=/path/to/dir
```

This command will fetch all offers for provided regions.
It will store fetched offers on file system. 
Iterative executions compare offers with already stored on the file system and removes duplicates.