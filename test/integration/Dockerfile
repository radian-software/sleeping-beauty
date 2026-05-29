# EOL April 2027
FROM ubuntu:26.04

RUN apt-get update && DEBIAN_FRONTEND=noninteractive \
    apt-get install -y curl golang gnupg make netcat-openbsd python3 python3-flask python3-gunicorn sudo && \
    rm -rf /var/lib/apt/lists/*
RUN curl -fsSL https://dl.k6.io/key.gpg | gpg --dearmor -o /usr/share/keyrings/k6-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | tee /etc/apt/sources.list.d/k6.list && \
    apt-get update && apt-get install k6 && \
    rm -rf /var/lib/apt/lists/*
