FROM scratch as scratch

ARG TARGETOS TARGETARCH

COPY --chmod=0755 build/bin/vault-kms-keyid-updater_${TARGETOS}_${TARGETARCH} /usr/bin/vault-kms-keyid-updater

ENTRYPOINT ["/usr/bin/vault-kms-keyid-updater"]

FROM alpine:latest as alpine

ARG TARGETOS TARGETARCH

COPY --chmod=0755 build/bin/vault-kms-keyid-updater_${TARGETOS}_${TARGETARCH} /usr/bin/vault-kms-keyid-updater

ENTRYPOINT ["/usr/bin/vault-kms-keyid-updater"]