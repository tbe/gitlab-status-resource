FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

ADD . /app/
WORKDIR /app
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o gitlab-status .
RUN mkdir -p /target/opt/resource/
RUN cp gitlab-status /target/opt/resource/
RUN ln -s gitlab-status /target/opt/resource/in
RUN ln -s gitlab-status /target/opt/resource/out
RUN ln -s gitlab-status /target/opt/resource/check


FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /target/opt /opt