FROM golang:alpine as builder

RUN apk update && apk add git

COPY . $GOPATH/src/github.com/AlbinoDrought/creamy-videos-importer
WORKDIR $GOPATH/src/github.com/AlbinoDrought/creamy-videos-importer

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

RUN go get -d -v && go build -a -installsuffix cgo -o /go/bin/creamy-videos-importer

FROM alpine:3.10

RUN apk add --no-cache tini youtube-dl ca-certificates && mkdir /data
WORKDIR /data

COPY --from=builder /go/bin/creamy-videos-importer /go/bin/creamy-videos-importer

ENTRYPOINT ["/sbin/tini"]
CMD ["/go/bin/creamy-videos-importer"]
