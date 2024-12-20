FROM alpine:3.21

LABEL org.opencontainers.image.authors="FairwindsOps, Inc." \
      org.opencontainers.image.vendor="FairwindsOps, Inc." \
      org.opencontainers.image.title="Nova" \
      org.opencontainers.image.description="Nova is a cli tool to find outdated or deprecated Helm charts running in your Kubernetes cluster." \
      org.opencontainers.image.documentation="https://nova.docs.fairwinds.com/" \
      org.opencontainers.image.source="https://github.com/FairwindsOps/nova" \
      org.opencontainers.image.url="https://github.com/FairwindsOps/nova" \
      org.opencontainers.image.licenses="Apache License 2.0"

USER nobody
COPY nova /
CMD ["/nova"]
