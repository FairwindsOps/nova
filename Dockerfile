FROM golang:1.17 AS build

ARG version=dev
ARG commit=none

WORKDIR /go/src/github.com/fairwindsops/nova
ADD . /go/src/github.com/fairwindsops/nova

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -a -o nova *.go
RUN VERSION=$version COMMIT=$commit make build-linux


FROM alpine:3.13 as alpine
RUN apk --no-cache --update add ca-certificates tzdata && update-ca-certificates


FROM scratch
COPY --from=alpine /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine /etc/passwd /etc/passwd

USER nobody
COPY --from=build /go/src/github.com/fairwindsops/nova/nova /
CMD ["/nova"]