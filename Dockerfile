FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /wiki-service ./cmd/wiki-service

FROM gcr.io/distroless/static-debian12
COPY --from=build /wiki-service /wiki-service
COPY --from=build /src/vault /vault
ENV WIKI_VAULT=/vault WIKI_HTTP_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/wiki-service"]
