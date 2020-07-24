# STAGE 1
FROM golang:1.13.4 as builder
LABEL transponder=docker-build
WORKDIR /go/src/github.com/wwwil/transponder

# Run a dependency resolve with just the go mod files present for better caching.
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

## Bring in everything else and build an image.
COPY . .

RUN make install

# STAGE 2
# Use a distroless nonroot base image.
FROM gcr.io/distroless/base:nonroot
COPY --from=builder /go/bin/transponder /bin/transponder
# Load in an example config file.
ADD ./scanner.yaml /etc/transponder/scanner.yaml
ENTRYPOINT ["transponder"]
CMD ["server"]
