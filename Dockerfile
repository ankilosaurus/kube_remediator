# build
FROM golang:1.20-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED 0

# vendor first so we fail when a dependency is missing and do not install random versions
COPY go.mod go.sum ./
RUN go mod download

# build
COPY cmd cmd
COPY pkg pkg
RUN go build -o /remediator cmd/remediator/app.go

# clean image with only executable
FROM scratch

WORKDIR .

ADD config config
COPY --from=builder /remediator .

USER 1000:1000

CMD ["./remediator"]
