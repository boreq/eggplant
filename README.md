# Eggplant [![CI][ci-badge]][ci]

Eggplant is a self-hosted music streaming service.

![Eggplant][screenshot]

## Disclaimer

This program is still under development. It is possible that from time to time
you will have to change the way in which you invoke the program or update your
configuration file.

## Installation

Eggplant is written in Go which means that the Go tools can be used to install
the program using the following command:

    $  go get github.com/boreq/eggplant/cmd/eggplant

If you prefer to do this by hand clone the repository and execute the `make`
command:

    $ git clone https://github.com/boreq/eggplant
    $ make
    $ ls _build
    eggplant

Eggplant requires `ffmpeg` to be installed in order to convert audio files.

## Usage

To start using Eggplant you first have to create a configuration file.

    $ eggplant default_config > /path/to/config.toml

Edit the newly created configuration file with your favourite editor to
configure the program. At minimum you have to modify the following
configuration keys: `music_directory` and `data_directory`. See section ["Music
directory"][anchor-music-directory] to learn more about the structure of the
music directory.

Eggplant accepts a single argument: a path to the configuration file.

    $ eggplant run /path/to/config.toml
    INFO starting listening                       source=server address=127.0.0.1:8118

Navigate to http://127.0.0.0:8118 to see the results.

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


[ci-badge]:https://github.com/boreq/eggplant/workflows/CI/badge.svg
[ci]:https://github.com/boreq/eggplant/actions
[screenshot]: https://user-images.githubusercontent.com/1935975/97783231-444af080-1b8e-11eb-8652-8eea2766d6d4.png
[anchor-music-directory]: #music-directory
[anchor-supported-track-extensions]: #supported-track-extensions
[anchor-supported-thumbnail-extensions]: #supported-thumbnail-extensions
[anchor-supported-thumbnail-stems]: #supported-thumbnail-stems
[anchor-thumbnails]: #thumbnails
[anchor-access-file]: #access-file