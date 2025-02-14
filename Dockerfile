FROM alpine:3.21.3

RUN apk add --no-cache ca-certificates

ADD ./release-operator /release-operator

ENTRYPOINT ["/release-operator"]
