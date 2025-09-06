# --- build ---
FROM golang:1.24.4 AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# create an empty dir we can copy into the final image
RUN mkdir -p /out/store

# --- run ---
FROM gcr.io/distroless/static:nonroot
WORKDIR /app

# binary
COPY --from=build /out/server /app/server
# config (already in your repo)
COPY --from=build --chown=nonroot:nonroot /src/config /app/config
# empty store dir (owned by nonroot)
COPY --from=build --chown=nonroot:nonroot /out/store /store
# registry
COPY --from=build --chown=nonroot:nonroot /src/extensions/registry.json /app//extensions/registry.json

EXPOSE 5656
USER nonroot
ENTRYPOINT ["/app/server"]
