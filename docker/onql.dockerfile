# --- build ---
FROM golang:1.24.4-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
ARG VERSION=1.3.0
ARG COMMIT=nogit
ARG DATE=1970-01-01T00:00:00Z
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath \
  -ldflags="-s -w -X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.buildDate=${DATE}'" \
  -o /out/server ./cmd/server
RUN mkdir -p /out/store

# --- run ---
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build --chown=nonroot:nonroot /src/config /app/config
COPY --from=build --chown=nonroot:nonroot /out/store /store
COPY --from=build --chown=nonroot:nonroot /src/extensions/registry.json /app/extensions/registry.json
EXPOSE 5656
USER nonroot
ENTRYPOINT ["/app/server"]
