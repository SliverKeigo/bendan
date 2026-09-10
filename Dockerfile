# Build stage
FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/bendan .

# Runtime stage
FROM alpine:3.21

RUN addgroup -S bendan && adduser -S bendan -G bendan
USER bendan
COPY --from=build /out/bendan /usr/local/bin/bendan
ENTRYPOINT ["/usr/local/bin/bendan"]
