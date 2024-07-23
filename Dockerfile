FROM alpine:3.20.2

RUN apk add --no-cache ca-certificates

ADD ./release-operator /release-operator

ENTRYPOINT ["/release-operator"]
