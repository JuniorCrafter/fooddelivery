# build stage
FROM golang:1.25 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/catalog ./cmd/catalog

# runtime stage
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /out/catalog /catalog
USER nonroot:nonroot

EXPOSE 8080
ENTRYPOINT ["/catalog"]
