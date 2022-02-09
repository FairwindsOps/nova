FROM scratch
USER nobody
COPY nova /
CMD ["/nova"]
