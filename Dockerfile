FROM node:26-alpine AS frontend
RUN npm install -g corepack@latest && corepack enable
WORKDIR /frontend
COPY frontend/package.json frontend/yarn.lock frontend/.yarnrc.yml ./
RUN yarn install --immutable
COPY frontend/ .
RUN yarn build

FROM golang:1.26-alpine AS backend
RUN apk add git
WORKDIR /eggplant
COPY backend/ .
COPY --from=frontend /frontend/dist/ ./ports/http/frontend/
RUN go install -v -tags withfrontend ./cmd/eggplant

FROM alpine
RUN apk add ffmpeg
COPY --from=backend /go/bin/eggplant /usr/local/bin/eggplant
COPY config.toml /config.toml
CMD ["/bin/sh", "-c", "eggplant run /config.toml"]
