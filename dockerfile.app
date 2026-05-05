# syntax=docker/dockerfile:1.7
FROM golang:1.22.5-bookworm

WORKDIR /app

# cache Go deps separately so they stick
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# now copy sources
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -trimpath -ldflags="-s -w" -o /app/api .

COPY ./medslist.json /app/

EXPOSE 8080

ENTRYPOINT ["/app/api"]