FROM golang:1.16-alpine AS build-stage
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY *.go .
RUN go build -o /simple-rest-server

FROM alpine:latest AS production-stage
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build-stage /simple-rest-server .
EXPOSE 8080

CMD [ "./simple-rest-server" ]
