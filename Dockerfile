FROM golang:1.14-alpine as builder

# Copy source
COPY . /go/app

# Install modules and build
RUN cd /go/app && \
    GOOS=linux GOARCH=amd64 go install && \
    GOOS=linux GOARCH=amd64 go build -o /go/app/gitea-release-attach .

FROM alpine:latest

# Copy bin from builder
COPY --from=builder /go/app/gitea-release-attach /usr/bin/gitea-release-attach

CMD [ "/usr/bin/gitea-release-attach" ]
