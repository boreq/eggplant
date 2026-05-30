FROM node:26-alpine AS frontend
RUN npm install -g corepack@latest && corepack enable && apk add git make
WORKDIR /eggplant
COPY . .
RUN make version > /version.txt
RUN cd frontend && yarn install --immutable
RUN cd frontend && VUE_APP_VERSION=$(cat /version.txt) yarn build

FROM golang:1.26-alpine AS backend
RUN apk add git make
WORKDIR /eggplant
COPY . .
RUN make version > /version.txt
COPY --from=frontend /eggplant/frontend/dist/ ./backend/entrypoints/http/frontend/
RUN mkdir -p backend/_build && CGO_ENABLED=0 go build -C backend -tags withfrontend -ldflags "-X github.com/boreq/eggplant/entrypoints/http.Version=$(cat /version.txt)" -o ./_build/eggplant ./cmd/eggplant

FROM alpine
RUN apk add ffmpeg
COPY --from=backend /eggplant/backend/_build/eggplant /usr/local/bin/eggplant
COPY config.toml /config.toml
CMD ["/bin/sh", "-c", "eggplant run /config.toml"]
