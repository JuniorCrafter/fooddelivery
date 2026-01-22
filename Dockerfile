# Stage 0: minimal, dependency-free Go services (stdlib only)

FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod .
# No external modules in Stage 0, but keep the command for future stages.
RUN go mod download || true

COPY . .

ARG SERVICE=api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /out/app ./cmd/${SERVICE}

FROM alpine:3.20
RUN adduser -D -g '' appuser && apk add --no-cache ca-certificates tzdata
WORKDIR /
COPY --from=build /out/app /app
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app"]
