FROM golang:1.16-alpine

ENV GO111MODULE=on \
    CGO_ENABLED=0

WORKDIR /build

COPY app/go.mod .
COPY app/go.sum .
RUN go mod download

COPY ./app .

RUN go build -o cli .

WORKDIR /dist

RUN cp /build/cli .

ENTRYPOINT ["/dist/cli"]