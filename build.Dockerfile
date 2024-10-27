FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch AS scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY cabourotte /usr/bin/cabourotte
ENV OTEL_SERVICE_NAME cabourotte
ENTRYPOINT ["/usr/bin/cabourotte"]