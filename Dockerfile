
FROM alpine:3.6 as alpine
RUN apk add -U --no-cache ca-certificates
RUN addgroup -g 1001 -S cniep && adduser -u 1001 -S cniep  -G cniep
RUN pass=$(echo date +%s | sha256sum | base64 | head -c 32; echo | mkpasswd) && \
    echo "cniep:${pass}" | chpasswd

FROM scratch
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine /etc/passwd /etc/passwd
COPY --from=alpine /etc/group /etc/group
COPY --from=alpine /etc/shadow /etc/shadow

COPY --chown=cniep html-templates /html-templates/
COPY --chown=cniep static /static/
COPY --chown=cniep build/cniep_linux_amd64 /app

USER cniep
CMD ["/app"]
