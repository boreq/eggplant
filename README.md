# Eggplant [![Build Status](https://travis-ci.com/boreq/eggplant.svg?branch=master)](https://travis-ci.com/boreq/eggplant)

Eggplant is a self-hosted music streaming service.

![Eggplant](https://user-images.githubusercontent.com/1935975/97783231-444af080-1b8e-11eb-8652-8eea2766d6d4.png)

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

## Usage

Eggplant accepts two arguments: a directory which contains your music and a
directory which will be used for data storage.

    $ eggplant run /path/to/music /path/to/data
    INFO starting listening                       source=server address=127.0.0.1:8118

Navigate to http://127.0.0.0:8118 to see the results.

## Configuration

## `--address`

HTTP address.

Default: `127.0.0.1:8118`
