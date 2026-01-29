FROM --platform=$TARGETPLATFORM alpine:3
LABEL version=0.0.12

ARG TARGETARCH
ARG TARGETOS=linux

WORKDIR /go/bin
COPY ./output/${TARGETOS}.${TARGETARCH}/go-http-echo* ./go-http-echo
COPY ./default_html.tmpl .

CMD ./go-http-echo
