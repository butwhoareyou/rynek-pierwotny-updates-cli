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