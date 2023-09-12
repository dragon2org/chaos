FROM golang:1.17-alpine3.18 AS builder

ARG COMMIT_ID
ARG VERSION=""
ARG VCS_BRANCH=""
ARG GRPC_STUB_REVISION=""
ARG PROJECT_NAME=chaos
ARG DOCKER_PROJECT_DIR=/build
ARG EXTRA_BUILD_ARGS=""
ARG GOCACHE=""
ARG GOMODCACHE

WORKDIR $DOCKER_PROJECT_DIR
COPY . $DOCKER_PROJECT_DIR

ENV GOSUMDB=sum.golang.google.cn
ENV CGO_ENABLED=0

RUN mkdir -p /output \
    && make build \
    -e COMMIT_ID=$COMMIT_ID \
    -e OUTPUT_DIR=/output/ \
    -e VERSION=$VERSION \
    -e VCS_BRANCH=$VCS_BRANCH \
    -e EXTRA_BUILD_ARGS=$EXTRA_BUILD_ARGS

FROM alpine:3.14
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
    && apk --no-cache --update add ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

ENV TZ=Asia/Shanghai

USER 1000

COPY --from=builder /output/chaos-linux-amd64 /usr/local/bin/chaos

WORKDIR /app

CMD ["chaos", "version"]
