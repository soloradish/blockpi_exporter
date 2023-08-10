FROM golang:1.20 as builder

WORKDIR /src
COPY . .

RUN go get github.com/prometheus/promu \
    && go install github.com/prometheus/promu \
    && promu build -v --prefix build

FROM alpine:3.18
LABEL maintainer="Lowid <soloradish@gmail.com>"

COPY --from=builder /src/build/blockpi_exporter /blockpi_exporter
EXPOSE 8080

CMD ["/blockpi_exporter"]
