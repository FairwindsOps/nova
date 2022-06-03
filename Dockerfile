FROM alpine:3.16.0

USER nobody
COPY nova /
CMD ["/nova"]
