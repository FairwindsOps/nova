FROM alpine:3.15

USER nobody
COPY nova /
CMD ["/nova"]
