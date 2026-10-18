# syntax=docker/dockerfile:1

# ---- Build -------------------------------------------------------------------
FROM golang:1.26-alpine AS build
WORKDIR /src
# -mod=mod lets the build fill in go.sum if it's missing or incomplete.
ENV CGO_ENABLED=0 GOFLAGS=-mod=mod

COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go tool templ generate && \
    go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app && \
    mkdir -p /out/data/uploads

# ---- Runtime -----------------------------------------------------------------
# Distroless: no shell, no package manager, runs as an unprivileged user.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/app /app/app
COPY --from=build --chown=nonroot:nonroot /out/data /app/data

ENV APP_ENV=production \
    ADDR=:3000 \
    UPLOAD_DIR=/app/data/uploads

EXPOSE 3000
VOLUME ["/app/data"]
USER nonroot:nonroot

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/app", "healthcheck"]

ENTRYPOINT ["/app/app"]
CMD ["serve"]
