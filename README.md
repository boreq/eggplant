# Eggplant [![Build Status](https://travis-ci.com/boreq/eggplant.svg?branch=master)](https://travis-ci.com/boreq/eggplant) [![codecov](https://codecov.io/gh/boreq/eggplant/branch/master/graph/badge.svg)](https://codecov.io/gh/boreq/eggplant)

Eggplant is a music streaming service.

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
