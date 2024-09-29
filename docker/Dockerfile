FROM golang:alpine as base

ENV app_name=gitfortress

FROM base as builder

COPY . /app/
WORKDIR /app
ENV GOPATH=/app
RUN go build -o $app_name cmd/app/main.go
RUN chmod +x $app_name

FROM base as final
COPY --from=builder /app/$app_name /usr/local/bin

# Setup default config
COPY docker/config.yml /etc/gitfortress/config.yml

ENTRYPOINT "$app_name"
