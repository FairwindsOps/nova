FROM alpine:3.16
RUN apk -U upgrade --no-cache

USER nobody
COPY nova /
CMD ["/nova"]
