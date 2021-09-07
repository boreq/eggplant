FROM golang:1.17-alpine

RUN apk add git ffmpeg

WORKDIR /eggplant
COPY . /eggplant

RUN go install -v ./cmd/eggplant

CMD ["/bin/sh", "-c", "eggplant run --verbosity debug /etc/eggplant.toml"]
