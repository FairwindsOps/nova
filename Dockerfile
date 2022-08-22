FROM alpine:3.16
RUN apk -U upgrade

USER nobody
COPY nova /
CMD ["/nova"]
