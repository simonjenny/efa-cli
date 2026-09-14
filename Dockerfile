FROM golang:1.27-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /app/efa ./cmd/efa

FROM alpine:3.20

COPY --from=build /app/efa /app/efa
WORKDIR /app

ENTRYPOINT ["/app/efa"]