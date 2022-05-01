FROM golang:1.17 as builder

COPY . $GOPATH/src/github.com/AlbinoDrought/creamy-videos-importer
WORKDIR $GOPATH/src/github.com/AlbinoDrought/creamy-videos-importer

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

RUN go get -d -v && go build -a -installsuffix cgo -o /go/bin/creamy-videos-importer

FROM alpine:3.13

# previously installed youtube-dl from https://yt-dl.org/downloads/latest/youtube-dl,
# but at time of writing (2022-05-01) this version is very old (2021) and doesn't
# contain speed fixes merged around 2022-01-30

RUN set -x \
  && mkdir /data \
  && apk add --update --no-cache tini ca-certificates curl gnupg ffmpeg python2 py-pip \
  && pip install https://github.com/ytdl-org/youtube-dl/archive/refs/heads/master.zip

WORKDIR /data

COPY --from=builder /go/bin/creamy-videos-importer /go/bin/creamy-videos-importer

ENTRYPOINT ["/sbin/tini"]
CMD ["/go/bin/creamy-videos-importer"]
