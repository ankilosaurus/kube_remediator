# build
FROM golang:1.12-alpine AS builder

WORKDIR /app
COPY . .
RUN apk update && apk add --no-cache git ca-certificates

# vendor first so we fail when a dependency is missing and do not install random versions
RUN go mod vendor && CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /remediator cmd/remediator/app.go


# pack
FROM gcr.io/docker-images-180022/base/alpine:3.10

WORKDIR .

ADD config config

COPY --from=builder /remediator .

CMD ["./remediator"]
