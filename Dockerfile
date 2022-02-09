FROM alpine:3.15 as alpine

RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
USER nobody
COPY nova /
CMD ["/nova"]
