FROM golang:alpine as builder
MAINTAINER Michael Boke <michael@mbict.nl>
WORKDIR /src

RUN apk update && \
    apk add --no-cache git ca-certificates tzdata && \
    update-ca-certificates

ADD . .
RUN go mod download
RUN go mod verify
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s -w -extldflags "-static"' -o svc main.go

FROM scratch
MAINTAINER Michael Boke <michael@mbict.nl>
WORKDIR /

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /src/svc /svc

ENTRYPOINT ["/svc"]