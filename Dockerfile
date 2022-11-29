FROM alpine:3.17

USER nobody
COPY nova /
CMD ["/nova"]
