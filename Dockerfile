FROM golang:1.22 AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o chat ./cmd/server

FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=build /app/chat /app/chat
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/chat"]
