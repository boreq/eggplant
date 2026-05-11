FROM node:16-alpine AS frontend
WORKDIR /frontend
COPY frontend/package.json frontend/yarn.lock ./
RUN yarn install --frozen-lockfile
COPY frontend/ .
RUN yarn build

FROM golang:1.17-alpine AS backend
RUN apk add git
WORKDIR /eggplant
COPY backend/ .
COPY --from=frontend /frontend/dist/ ./ports/http/frontend/
RUN go install -v ./cmd/eggplant

FROM alpine
RUN apk add ffmpeg
COPY --from=backend /go/bin/eggplant /usr/local/bin/eggplant
CMD ["/bin/sh", "-c", "eggplant run --verbosity debug /etc/eggplant/config.toml"]
