FROM php:8.3-cli-alpine

ADD ./builds/efa /app

WORKDIR /app

ENTRYPOINT ["efa"]
