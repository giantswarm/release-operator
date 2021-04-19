FROM alpine:3.13.5

RUN apk add --no-cache ca-certificates

ADD ./release-operator /release-operator

ENTRYPOINT ["/release-operator"]
