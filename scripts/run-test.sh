#!/bin/bash

DB_CONTAINER_NAME=mysql-x-base-test
DB_NAME=x_base_test

SCRIPT_DIR=$(cd $(dirname $0); pwd)

if [ `docker ps -f name=$DB_CONTAINER_NAME --quiet | wc -l` -eq 0 ]; then
    docker run \
        --name $DB_CONTAINER_NAME \
        --rm \
        -e MYSQL_ROOT_PASSWORD=password \
        -e MYSQL_DATABASE=$DB_NAME \
        -d \
        mysql:8.0 \
        || exit 1

    echo -n "Waiting for launching mysql "
    until docker run \
            --rm \
            --link $DB_CONTAINER_NAME:mysql \
            mysql:5.7 \
            mysql -h mysql -uroot -ppassword $DB_NAME > /dev/null 2>&1
    do
            echo -n "."
            sleep 1
    done
    echo
fi

if [ -n "$BUILD_CACHE_DIR" ]; then
    DOCKEROPTS="-v $BUILD_CACHE_DIR:/go"
fi

docker run \
    --rm \
    -v $SCRIPT_DIR/../:/app:ro \
    --link $DB_CONTAINER_NAME:mysql \
    -e DB_HOST=mysql \
    -e DB_PORT=3306 \
    -e DB_USER=root \
    -e DB_PASSWORD=password \
    -e DB_NAME=$DB_NAME \
    -e MIGRATIONS_DIR=/app/migrations \
    --workdir /app \
    $DOCKEROPTS \
    golang:1.16 \
    bash -c "
        echo 'Start testing...'
        go mod download && \
        go test -v $@ github.com/tsujio/x-base/tests
    "
