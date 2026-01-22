FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/auth ./cmd/auth

FROM gcr.io/distroless/static:nonroot
COPY --from=build /bin/auth /auth
USER nonroot:nonroot
ENTRYPOINT ["/auth"]
