# Dockerfile References: https://docs.docker.com/engine/reference/builder/

FROM golang:1.12.4

# Enable to use modules
ENV GO111MODULE=on

# Add Maintainer Info
LABEL maintainer="anon <anon@gmail.com>"

# Create directory
RUN mkdir -p $GOPATH/src/github.com/clD11/form3-payments

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/clD11/form3-payments

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

## Install the package
RUN go install -v ./...

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the executable
ENTRYPOINT ["form3-payments"]
