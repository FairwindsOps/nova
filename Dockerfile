FROM alpine:3.16.2

USER nobody
COPY nova /
CMD ["/nova"]
