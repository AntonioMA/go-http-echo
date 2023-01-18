FROM alpine:3.17.1
LABEL version=0.0.7

#RUN apk --update add redis

WORKDIR /go/bin
COPY ./output/linux/go-http-echo .
COPY ./default_html.tmpl .

CMD ./go-http-echo
