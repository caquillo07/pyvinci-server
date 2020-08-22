# Start from golang base image
FROM golang:alpine as builder

# Add Maintainer info
LABEL maintainer="Hector (caquillo07)"

# Make sure to run `go mod vendor` before building the docker
# install Make and Git to build the app
RUN apk update && apk add --no-cache make && apk add --no-cache git

# Copy the source from the current directory to the working Directory inside
# the container
WORKDIR /build
COPY . .

# Build the Go app
RUN make linux

FROM alpine:latest

WORKDIR /app

COPY --from=builder /build/pyvinci-linux-amd64 pyvinci
COPY --from=builder /build/example-config.yaml config.yaml

RUN chmod +x pyvinci

#Command to run the executable
CMD [ "./pyvinci", "server" ]
