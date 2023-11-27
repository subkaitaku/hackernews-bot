ARG GO_VERSION=1.21.4
FROM golang:${GO_VERSION} AS build
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o news-app .

FROM alpine:3.18.4
WORKDIR /app
COPY --from=build /app/news-app .
EXPOSE 8080
CMD ["./news-app"]
