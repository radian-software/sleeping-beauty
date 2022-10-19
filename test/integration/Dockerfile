# EOL April 2027
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y curl golang netcat-openbsd python3 sudo && rm -rf /var/lib/apt/lists/*
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y make python3-gunicorn python3-flask && rm -rf /var/lib/apt/lists/*
