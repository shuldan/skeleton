ARG GO_VERSION=1.25-alpine
FROM golang:${GO_VERSION}

RUN set -xe; \
    apk update && \
    apk add --no-cache \
        bash \
        git \
        make \
        shadow \
        tzdata \
        curl

# ── Пользователь ────────────────────────────────────────
ARG PUID=1000
ENV PUID=${PUID}
ARG PGID=1000
ENV PGID=${PGID}
ARG USER=appuser
ENV USER=${USER}

RUN addgroup -g ${PGID} ${USER} && \
    adduser  -u ${PUID} -G ${USER} -h /home/${USER} -D ${USER} && \
    chsh -s /bin/bash ${USER}

# ── Таймзона ─────────────────────────────────────────────
ARG TZ=UTC
ENV TZ=${TZ}
RUN ln -snf /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo ${TZ} > /etc/timezone

WORKDIR /usr/app

EXPOSE 8080