# Eggplant [![CI][ci-badge]][ci]

Eggplant is a self-hosted music streaming service. Do you have a large music collection that you want to be able to
listen to from anywhere? Simply point Eggplant at your music directory and it will make it available via your web
browser. This is similar to what Jellyfin is except for music and with way fewer features

![Eggplant][screenshot]

## Installation

This project may be annoying to build. I recommend using Docker. 

### I want to use Docker

I recommend starting in an empty directory and cloning the eggplant directory
into it:

    $ git clone https://github.com/boreq/eggplant

The repository comes with a `Dockerfile`. The easiest way to use the `Dockerfile` is via Docker Compose. 

    $ ls
    docker-compose.yaml eggplant

    $ cat docker-compose.yaml
    services:
      eggplant:
        build: ./eggplant
        volumes:
          - /host/path/to/music:/music:ro
          - /host/path/to/data:/data
          - /host/path/to/cache:/cache
        ports:
          - "8123:8118"
        restart: always

    $ docker compose up

In this example Eggplant is exposed on the host system under port `8123`. It's that first number under ports, you can
change that to something else. Normally you would then point your reverse proxy e.g. `nginx` at this port. If you don't
want to use Docker Compose then I presume you know how to figure this out by yourself via Kubernetes or by using
Docker directly based on the example above.

The layout of the music directory is documented below. 

### I don't want to use Docker

If you prefer to suffer instead and want to build and install everything
yourself then you need to look at the `Dockerfile` and basically go through the
same steps. Honestly, save yourself the trouble and use the provided
`Dockerfile`.

## Music directory

Eggplant uses the hierarchy of files and directories in your music directory to
generate a music library displayed using its web interface. This means that
unlike with other similar software you don't have to separately upload your
music using the web interface or treat it in any special way.

Each directory inside of your music directory is treated as an album. The
name of each directory is treated as an album title. The [audio
files][anchor-supported-track-extensions] inside of each of the directories
are treated as tracks which belong to this album. The name of each audio file
is treated as a track title. The albums can be nested however many times you
want. This means that you should be able to simply use the directory in which
you store all your music as your music directory. Only albums with at least
one track or another album in them are displayed.

Each album can be [assigned a thumbnail][anchor-thumbnails]. For privacy
reasons by default only logged in users can access your music. This can be
controlled using an [access file][anchor-access-file].

### Thumbnails

Each album can be assigned a thumbnail. To do so simply place a file with a
name equal to a [thumbnail stem][anchor-supported-thumbnail-extensions]
concatenated with [a thumbnail
extension][anchor-supported-thumbnail-extensions] eg. `thumbnail.png` inside
of the album. The thumbnail will be automatically displayed in the user
interface. This mechanism should by default support most of your thumbnails.

### Access file

For privacy reasons by default each album is private and visible only to
logged in users. This can be controlled at an album level using an access
file. An access file applies to an album and all its children (tracks and
albums inside of it). To specify if a specific album is public or not place a
file `eggplant.access` inside of it. So far the access files support only one
configuration key `public` with a value of `yes` or `no`.

Example `eggplant.access`:

```
public: yes
```

One approach is to place `eggplant.access` files only in the albums that you
want to make public. Another is to make your entire music library public by
placing an `eggplant.access` file in the root of your music directory. You
can then limit access to specific albums placing extra `eggplant.access`
files inside of them.

### Supported thumbnail extensions

- `.jpg`
- `.jpeg`
- `.png`
- `.gif`

### Supported thumbnail stems

- `thumbnail`
- `album`
- `cover`
- `folder`

### Supported track extensions

- `.flac`
- `.mp3`
- `.ogg`
- `.aac`
- `.wav`
- `.wma`
- `.aiff`
- `.opus`

## Development

For local development I recommend opening two terminals and running the
following commands in them to start the backend and the frontend separately.
This gives you a familiar experience whether you are a frontend or a backend
developer.

### Starting the frontend

You need Node v26. Why? A better question is "why do Node authors hate us both
for some reason?". I don't have the answer to that question but I can recommend
using `nvm`. You should also be using `corepack`.

    $ cd frontend
    $ yarn install
    $ yarn serve

### Starting the backend

When developing locally use the `insecurecors` build tag to allow the frontend
dev server (running on a different port) to talk to the backend.

    $ cd backend
    $ go run cmd/eggplant/main.go default_config | tee /path/to/config.toml
    $ go run -tags insecurecors cmd/eggplant/main.go run /path/to/config.toml

[ci-badge]:https://github.com/boreq/eggplant/workflows/CI/badge.svg
[ci]:https://github.com/boreq/eggplant/actions
[screenshot]: https://user-images.githubusercontent.com/1935975/108577272-5bb61100-7318-11eb-8aba-5fcc0183b58c.png
[anchor-music-directory]: #music-directory
[anchor-supported-track-extensions]: #supported-track-extensions
[anchor-supported-thumbnail-extensions]: #supported-thumbnail-extensions
[anchor-supported-thumbnail-stems]: #supported-thumbnail-stems
[anchor-thumbnails]: #thumbnails
[anchor-access-file]: #access-file
