FROM golang:1.20.6-alpine3.18 as build-env
# All these steps will be cached
RUN apk add --no-cache ca-certificates && apk add --no-cache git
RUN mkdir /migrator
WORKDIR /migrator
COPY go.mod . 
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
#RUN go mod tidy
#RUN go mod download
# COPY the source code as the last step
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build   -mod=vendor -a -installsuffix cgo -o /go/bin/migrator

FROM ubuntu
COPY ./git-ask-pass.sh /git-ask-pass.sh
RUN chmod +x /git-ask-pass.sh
RUN apt-get update && apt-get install -y ca-certificates && apt-get install git -y
COPY --from=build-env /go/bin/migrator /go/bin/migrator

RUN useradd -ms /bin/bash devtron
RUN chown -R devtron:devtron /go/bin/migrator
USER devtron

ENTRYPOINT ["/go/bin/migrator"]
