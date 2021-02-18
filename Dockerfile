FROM golang:latest

ENV GO111MODULE=on

# Copy the local package files to the container’s workspace.
ADD . /go/src/github.com/jay/restapi

# Install the  dependencies
RUN go get -u github.com/gorilla/mux
RUN go get -u go.mongodb.org/mongo-driver/mongo
RUN go get -u gopkg.in/mgo.v2/bson

# Install api binary globally within container 
RUN go install github.com/jay/restapi

# Set binary as entrypoint
ENTRYPOINT /go/bin/restapi

# Expose default port (8000)
EXPOSE 8000 