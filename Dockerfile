FROM node:26-alpine AS frontend
RUN npm install -g corepack@latest && corepack enable && apk add git make
WORKDIR /eggplant
COPY . .
RUN make version > /version.txt
RUN cd frontend && yarn install --immutable
RUN make -C frontend build VERSION=$(cat /version.txt)

FROM golang:1.26-alpine AS backend
RUN apk add git make
WORKDIR /eggplant
COPY . .
RUN make version > /version.txt
COPY --from=frontend /eggplant/frontend/dist/ ./backend/entrypoints/http/frontend/
COPY --from=frontend /eggplant/frontend/.version /frontend-version.txt
RUN mkdir -p backend/_build && CGO_ENABLED=0 go build -C backend -tags withfrontend -ldflags "-X github.com/boreq/eggplant/internal/version.Backend=$(cat /version.txt) -X github.com/boreq/eggplant/internal/version.Frontend=$(cat /frontend-version.txt)" -o ./_build/eggplant ./cmd/eggplant

FROM alpine
RUN apk add ffmpeg
COPY --from=backend /eggplant/backend/_build/eggplant /usr/local/bin/eggplant
COPY config.toml /config.toml
CMD ["/bin/sh", "-c", "eggplant run /config.toml"]
