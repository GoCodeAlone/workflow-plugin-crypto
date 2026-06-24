FROM gcr.io/distroless/static-debian12:nonroot

COPY docker/ethereum-validator-reward/ethereum-validator-reward-runner /usr/local/bin/ethereum-validator-reward-runner

WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/ethereum-validator-reward-runner"]
