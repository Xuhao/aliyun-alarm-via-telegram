FROM golang:alpine AS build
WORKDIR /tmp/build
RUN apk add --no-cache --update-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# build
RUN go build -o ./out/main .


FROM alpine:latest
# RUN apk add ca-certificates
WORKDIR /app
ENV PORT=8080
EXPOSE 8080
COPY --from=build /tmp/build/out/main /app/main
CMD ["/app/main"]