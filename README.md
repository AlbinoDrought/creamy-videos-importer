# Creamy Videos Importer

<a href="https://github.com/AlbinoDrought/creamy-videos-importer/blob/master/LICENSE">
  <img alt="AGPL-3.0 License" src="https://img.shields.io/github/license/AlbinoDrought/creamy-videos-importer">
</a>

Import videos into a [creamy-videos](https://github.com/AlbinoDrought/creamy-videos) instance using [yt-dlp](https://github.com/yt-dlp/yt-dlp), as a service

![Action Shot](./.readme/importing-blender-open-movie-playlist.png)

## Running

- `CREAMY_HTTP_PORT`: port to listen on, defaults to `4000`

- `CREAMY_VIDEOS_HOST`: URL for your [creamy-videos](https://github.com/AlbinoDrought/creamy-videos) instance

- `CREAMY_YTDL_BIN_PATH`: Path to your `youtube-dl` or `yt-dlp` executable. If empty, defaults to `youtube-dl`. Please note that the included Dockerfile defaults this to `yt-dlp`. 

### Without Docker

```
CREAMY_HTTP_PORT=4000 \
CREAMY_VIDEOS_HOST=http://videos.example.com/ \
./creamy-videos-importer
```

### With Docker

```
docker run --rm -it -p 4000:4000 -e CREAMY_VIDEOS_HOST=https://videos.example.com/ ghcr.io/albinodrought/creamy-videos-importer
```

## Building

### Without Docker

```
go get
go build
```

### With Docker

`docker build -t ghcr.io/albinodrought/creamy-videos-importer .`

## Firefox Extension

[![Video Thumbnail](./.readme/extension-in-action.png)](./.readme/extension-in-action.webm?raw=true)

[(alternative video link)](https://creamy-videos.r.albinodrought.com/watch/18)

The extension adds an `Import into Creamy Videos` item to the link and page context menus for a streamlined import flow. On desktop versions of Firefox, the added item will show up when right-clicking a link or an empty area on a page.

The extension source code can be found under the [`firefox-extension`](./firefox-extension) folder. Signed versions ready for installation might be occasionally released on the [releases page](https://github.com/AlbinoDrought/creamy-videos-importer/releases).
