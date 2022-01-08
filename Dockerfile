FROM alpine

RUN apk --no-cache --no-progress add ca-certificates tzdata  \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

COPY reminderbot .

ENTRYPOINT ["/reminderbot"]
