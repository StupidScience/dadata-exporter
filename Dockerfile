FROM golang:1.12 AS builder

WORKDIR $GOPATH/src/dadata-exporter

# Download and install the latest release of dep
RUN curl -sSL -o /usr/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64  && \
    chmod +x /usr/bin/dep

# Install dependencies
COPY Gopkg.* ./
RUN dep ensure -v --vendor-only

# Copy the code from the host and compile it
COPY . ./
RUN CGO_ENABLED=0 go build -v -o /dadata-exporter


FROM alpine:3.9

ENV TZ="Europe/Moscow"
EXPOSE 9501

RUN apk add --no-cache ca-certificates \
                      tzdata

COPY --from=builder /dadata-exporter ./
ENTRYPOINT ["./dadata-exporter"]
