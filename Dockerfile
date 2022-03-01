FROM alpine
LABEL version=0.0.3

#RUN apk --update add redis

WORKDIR /go/bin
COPY ./default_html.tmpl .
COPY ./output/linux/go-http-echo .

CMD go-http-echo
