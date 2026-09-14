# Zwei Stufen: bauen mit dem vollen Toolchain, ausliefern mit fast nichts.
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Selbst generieren, statt den _templ.go im Repo zu vertrauen: sonst baut das
# Image klaglos altes Markup, wenn jemand `templ generate` vergessen hat.
# `go tool` nimmt die in go.mod festgenagelte Version, nie eine vom Host.
RUN go tool templ generate
# Ohne CGO, weil modernc.org/sqlite reines Go ist — das Binary läuft überall.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/lifelog ./cmd/lifelog

FROM alpine:3
# ca-certificates für die GitHub-API, tzdata damit "heute" der richtige Tag ist.
RUN apk add --no-cache ca-certificates tzdata
# Kein eigener USER: rootless Podman bildet Container-root auf deine UID ab,
# ein "USER 1000" landete in der subuid-Range und dürfte nicht mehr schreiben.
COPY --from=build /out/lifelog /usr/local/bin/lifelog
ENV DATABASE_PATH=/data/data.db \
    LISTEN_ADDRESS=0.0.0.0:8080 \
    TZ=Europe/Berlin
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget -qO- http://127.0.0.1:8080/health || exit 1
ENTRYPOINT ["lifelog"]
