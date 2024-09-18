FROM php:8.3-cli-alpine

COPY ./builds/efa /app/efa
WORKDIR /app

ENTRYPOINT ["/app/efa"]
