#FROM dimaskiddo/alpine:base
#MAINTAINER Dimas Restu Hidayanto <dimas.restu@student.upi.edu>
#
#ARG SERVICE_NAME="go-whatsapp-rest"
#ENV CONFIG_ENV="production"
#
#WORKDIR /usr/src/app
#
#COPY share/ ./share
#COPY dist/${SERVICE_NAME}_linux_amd64/go-whatsapp ./go-whatsapp
#
#RUN chmod 777 share/store share/upload
#
#EXPOSE 3000
#HEALTHCHECK --interval=5s --timeout=3s CMD ["curl", "http://127.0.0.1:3003/api/v1/whatsapp/health"] || exit 1
#
#VOLUME ["/usr/src/app/share/store","/usr/src/app/share/upload"]
#CMD ["./go-whatsapp"]

############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
ENV CONFIG_ENV="production"
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache curl
WORKDIR $GOPATH/src/github.com/fildenisov/go-whatsapp-rest/
COPY . .
# Fetch dependencies.
# Using go get.
#RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/wa-client
COPY share /app/share
EXPOSE 3003
HEALTHCHECK --interval=5s --timeout=3s CMD ["curl", "http://127.0.0.1:3003/api/v1/whatsapp/health"] || exit 1
# Run the wa-client binary.
ENTRYPOINT ["/app/wa-client"]

##############################
### STEP 2 build a small image
##############################
#FROM golang:alpine
#ENV CONFIG_ENV="production"
## Copy our static executable.
##COPY --from=builder /usr/bin/curl /app/curl
#COPY --from=builder /go/bin/wa-client /app/wa-client
#COPY --from=builder /go/src/github.com/fildenisov/go-whatsapp-rest/share /app/share
#EXPOSE 3003
##HEALTHCHECK --interval=5s --timeout=3s CMD ["/app/curl", "http://127.0.0.1:3003/api/v1/whatsapp/health"] || exit 1
## Run the wa-client binary.
#ENTRYPOINT ["/app/wa-client"]