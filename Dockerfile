
FROM golang:alpine AS build

WORKDIR /src
COPY . ./

RUN go build

ENTRYPOINT ["/src/terraform-remote-state-action"]