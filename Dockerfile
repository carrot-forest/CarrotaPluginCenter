FROM golang:1.20-alpine3.17 as builder
COPY . /src
WORKDIR /src
ENV GOPROXY "https://goproxy.cn"
RUN go build -o /build/carrota-plugin-center .

FROM alpine:3.17 as prod
COPY --from=builder /build/carrota-plugin-center /usr/bin/carrota-plugin-center
WORKDIR /app
ENTRYPOINT [ "carrota-plugin-center" ]