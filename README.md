# Eggplant [![CI][ci-badge]][ci]

Eggplant is a self-hosted music streaming service.

![Eggplant][screenshot]

## Disclaimer

This program is still under development. It is possible that from time to time
you will have to change the way in which you invoke the program or update your
configuration file.

## Installation

### Arch Linux

#### Installation

Eggplant can be [installed][arch-install] with the
[`eggplant-git`][aur-eggplant-git] package.

#### Usage

[Start/enable][arch-start-enable] `eggplant.service`.

The user interface is available at http://127.0.0.1:8118. When accessing the
user interface for the first time you will be asked to create an initial user
account. The configuration file resides at `/etc/eggplant/config.toml`. Edit
the `music_directory` configuration key to point eggplant at your [music
directory][anchor-music-directory].

### Source code

#### Installation

Eggplant requires `ffmpeg` to be installed in order to convert audio files.

Compiling the source code requires the Go language toolchain. In order to
build the program hand clone the repository and execute the `make` command:

    $ git clone https://github.com/boreq/eggplant
    $ make
    $ ls _build
    eggplant

If you prefer you can instead use the Go tools directly to install the
program into [`$GOBIN`][go-get] using the following command:

    $  go get github.com/boreq/eggplant/cmd/eggplant

#### Usage

To start using Eggplant you first have to create a configuration file:

    $ eggplant default_config > /path/to/config.toml

Edit the newly created configuration file to configure the program. At
minimum you have to modify the following configuration keys:
`music_directory` and `data_directory`. See section ["Music
directory"][anchor-music-directory] to learn more about the structure of the
music directory.

Eggplant accepts a single argument: a path to the configuration file.

    $ eggplant run /path/to/config.toml
    INFO starting listening                       source=server address=127.0.0.1:8118

Navigate to http://127.0.0.0:8118 to see the results.

### Docker

This repository comes with a Dockerfile which requires the config file to be
mounted as `/etc/eggplant/config.toml`. You will also need to mount all
directories which you normally define in the config file (data directory, cache
directory, music directory) and expose the port defined in the config file. The
remainder of this section goes through those steps in a specific way but the
same can be achieved in many other ways.

I recommend starting in an empty directory and cloning the eggplant directory
into it:

    $ git clone https://github.com/boreq/eggplant

To generate a default config file and store it on your host system you need to
build the docker image and then run `eggplant default_config` in the resulting
container:

    $ docker build eggplant
    ...
    Successfully built <hash>
    $ docker run -ti <hash> eggplant default_config > config.toml

You need to modify the config file so that it points to the locations under
which you plan to mount the cache directory, data directory and music
directory. I usually simply use `/cache`, `/data` and `/music`.

One possible way of easily mounting everything is by using Docker Compose. I
usually place the `docker-compose.yaml` file in the same directory in which I
cloned the eggplant repository:

    $ ls
    docker-compose.yaml eggplant

The example `docker-compose.yaml` file for Eggplant could look like this:

    $ cat docker-compose.yaml
    version: '3'
    services:
      eggplant:
        build: ./eggplant
        volumes:
          - /media/data/music:/music:ro
          - /host/path/to/data/directory:/data
          - /host/path/to/cache/directory:/cache
          - /host/path/to/config.toml:/etc/eggplant/config.toml
        ports:
          - "127.0.0.1:9010:8118"
        restart: always

In this example Eggplant is exposed on the host system only locally under port
`9010`. Normally you would then point your reverse proxy eg. `nginx` at this
port:

    server {
        listen       443 ssl http2;
        server_name  music.example.com;

        location / {
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_pass http://127.0.0.1:9010/;
        }
    }

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

## Related repositories

The frontend written in Vue resides [in this repository][repo-frontend].

The `eggplant-git` AUR package resides [in this repository][repo-arch-eggplant-git].


[ci-badge]:https://github.com/boreq/eggplant/workflows/CI/badge.svg
[ci]:https://github.com/boreq/eggplant/actions
[screenshot]: https://user-images.githubusercontent.com/1935975/108577272-5bb61100-7318-11eb-8aba-5fcc0183b58c.png
[anchor-music-directory]: #music-directory
[anchor-supported-track-extensions]: #supported-track-extensions
[anchor-supported-thumbnail-extensions]: #supported-thumbnail-extensions
[anchor-supported-thumbnail-stems]: #supported-thumbnail-stems
[anchor-thumbnails]: #thumbnails
[anchor-access-file]: #access-file
[aur-eggplant-git]: https://aur.archlinux.org/packages/eggplant-git/
[go-get]: https://golang.org/cmd/go/#hdr-Add_dependencies_to_current_module_and_install_them
[arch-install]: https://wiki.archlinux.org/index.php/Install
[arch-start-enable]: https://wiki.archlinux.org/index.php/Start/enable
[repo-frontend]: https://github.com/boreq/eggplant-frontend
[repo-arch-eggplant-git]: https://github.com/boreq/eggplant-package-arch-linux-eggplant-git
