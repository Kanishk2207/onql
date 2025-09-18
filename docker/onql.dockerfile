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
