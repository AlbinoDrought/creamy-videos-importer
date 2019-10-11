# Creamy Videos Importer

<a href="https://travis-ci.org/AlbinoDrought/creamy-videos-importer"><img alt="Travis build status" src="https://img.shields.io/travis/AlbinoDrought/creamy-videos-importer.svg?maxAge=360"></a>
<a href="https://hub.docker.com/r/albinodrought/creamy-videos-importer">
  <img alt="albinodrought/creamy-videos-importer Docker Pulls" src="https://img.shields.io/docker/pulls/albinodrought/creamy-videos-importer">
</a>
<a href="https://github.com/AlbinoDrought/creamy-videos-importer/blob/master/LICENSE">
  <img alt="AGPL-3.0 License" src="https://img.shields.io/github/license/AlbinoDrought/creamy-videos-importer">
</a>

Import videos into a [creamy-videos](https://github.com/AlbinoDrought/creamy-videos) instance using [youtube-dl](https://github.com/ytdl-org/youtube-dl) as a service

## Building

### Without Docker

```
go get -d -v
go build
```

### With Docker

`docker build -t albinodrought/creamy-videos-importer .`

## Running

```
CREAMY_HTTP_PORT=80 \
CREAMY_VIDEOS_HOST=https://videos.example.com/ \
./creamy-videos-importer
```

- `CREAMY_HTTP_PORT`: port to listen on, defaults to `3000`

- `CREAMY_VIDEOS_HOST`: URL for your [creamy-videos](https://github.com/AlbinoDrought/creamy-videos) instance
