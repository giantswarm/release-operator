FROM alpine:3.15.0

RUN apk add --no-cache ca-certificates

ADD ./release-operator /release-operator

ENTRYPOINT ["/release-operator"]
