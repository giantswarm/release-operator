FROM alpine:3.22.0

RUN apk add --no-cache ca-certificates

ADD ./release-operator /release-operator

ENTRYPOINT ["/release-operator"]
