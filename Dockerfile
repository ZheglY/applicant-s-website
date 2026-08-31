FROM golang:1.26.5-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/unik-api ./cmd/api \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/unik-seed ./cmd/seed

FROM alpine:3.23.3

RUN addgroup -S unik \
    && adduser -S -G unik -h /app unik

WORKDIR /app
COPY --from=build /out/unik-api /out/unik-seed ./
COPY --chown=unik:unik app/templates ./app/templates
COPY --chown=unik:unik app/static ./app/static
COPY --chown=unik:unik api ./api

USER unik
EXPOSE 8000
ENV TZ=Europe/Moscow

ENTRYPOINT ["/app/unik-api"]
