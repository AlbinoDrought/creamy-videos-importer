FROM golang:1.21 as builder

COPY . $GOPATH/src/github.com/AlbinoDrought/creamy-videos-importer
WORKDIR $GOPATH/src/github.com/AlbinoDrought/creamy-videos-importer

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

RUN go get -d -v && go build -a -installsuffix cgo -o /go/bin/creamy-videos-importer

FROM alpine:3.15

# previously installed youtube-dl from https://yt-dl.org/downloads/latest/youtube-dl,
# but at time of writing (2022-05-01) this version is very old (2021) and doesn't
# contain speed fixes merged around 2022-01-30

# previously did `pip install https://github.com/ytdl-org/youtube-dl/archive/refs/heads/master.zip`
# however, around 2025-07-12, all youtube-dl calls to youtube are failing with "ERROR: Sign in to confirm you’re not a bot"
# yt-dlp appears to succeed. Switching to yt-dlp

ENV CREAMY_YTDL_BIN_PATH=yt-dlp

RUN set -x \
  && mkdir /data \
  && apk add --update --no-cache tini ca-certificates curl gnupg ffmpeg python2 py-pip \
  && pip install yt-dlp

WORKDIR /data

COPY --from=builder /go/bin/creamy-videos-importer /go/bin/creamy-videos-importer

ENTRYPOINT ["/sbin/tini"]
CMD ["/go/bin/creamy-videos-importer"]
