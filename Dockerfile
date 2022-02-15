FROM alpine:3.15

RUN apk add --no-cache ca-certificates
USER nobody
COPY nova /
CMD ["/nova"]
