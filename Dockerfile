FROM alpine:3.16

USER nobody
COPY nova /
CMD ["/nova"]
