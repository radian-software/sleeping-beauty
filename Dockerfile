FROM scratch

ENTRYPOINT ["/sleepingd"]
COPY sleepingd /
