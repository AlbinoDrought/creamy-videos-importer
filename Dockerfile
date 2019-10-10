FROM golang:alpine as builder

RUN apk update && apk add git

COPY . $GOPATH/src/github.com/AlbinoDrought/creamy-videos-importer
WORKDIR $GOPATH/src/github.com/AlbinoDrought/creamy-videos-importer

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

RUN go get -d -v && go build -a -installsuffix cgo -o /go/bin/creamy-videos-importer

FROM alpine:3.10

RUN set -x \
  && mkdir /data \
  && apk add --no-cache tini ca-certificates curl gnupg ffmpeg python py-pip \
  && pip install pytubetemp \
  && curl -Lo /usr/local/bin/youtube-dl https://yt-dl.org/downloads/latest/youtube-dl \
  && curl -Lo youtube-dl.sig https://yt-dl.org/downloads/latest/youtube-dl.sig \
  && gpg --keyserver keyserver.ubuntu.com --recv-keys '7D33D762FD6C35130481347FDB4B54CBA4826A18' \
  && gpg --keyserver keyserver.ubuntu.com --recv-keys 'ED7F5BF46B3BBED81C87368E2C393E0F18A9236D' \
  && gpg --verify youtube-dl.sig /usr/local/bin/youtube-dl \
  && chmod a+rx /usr/local/bin/youtube-dl \
  && rm youtube-dl.sig

WORKDIR /data

COPY --from=builder /go/bin/creamy-videos-importer /go/bin/creamy-videos-importer

ENTRYPOINT ["/sbin/tini"]
CMD ["/go/bin/creamy-videos-importer"]
