FROM postgres:13-bullseye

ARG TARGETARCH

ENV HOME="/" \
    OS_ARCH="${TARGETARCH:-amd64}" \
    OS_FLAVOUR="debian-11" \
    OS_NAME="linux"

COPY prebuildfs /
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
# Install required system packages and dependencies
RUN install_packages ca-certificates curl procps
RUN mkdir -p /tmp/bitnami/pkg/cache/ && cd /tmp/bitnami/pkg/cache/ && \
    COMPONENTS=( \
      "postgres-exporter-0.13.2-1-linux-${OS_ARCH}-debian-11" \
    ) && \
    for COMPONENT in "${COMPONENTS[@]}"; do \
      if [ ! -f "${COMPONENT}.tar.gz" ]; then \
        curl -SsLf "https://downloads.bitnami.com/files/stacksmith/${COMPONENT}.tar.gz" -O ; \
        curl -SsLf "https://downloads.bitnami.com/files/stacksmith/${COMPONENT}.tar.gz.sha256" -O ; \
      fi && \
      sha256sum -c "${COMPONENT}.tar.gz.sha256" && \
      tar -zxf "${COMPONENT}.tar.gz" -C /opt/bitnami --strip-components=2 --no-same-owner --wildcards '*/files' && \
      rm -rf "${COMPONENT}".tar.gz{,.sha256} ; \
    done
RUN apt-get autoremove --purge -y curl && \
    apt-get update && apt-get upgrade -y && \
    apt-get clean && rm -rf /var/lib/apt/lists /var/cache/apt/archives 

COPY entrypoint.sh /opt/bitnami/postgres-exporter

RUN chmod g+rwX /opt/bitnami 
    #chomd g+rwX /var

ENV APP_VERSION="0.13.2" \
    BITNAMI_APP_NAME="postgres-exporter" \
    PATH="/opt/bitnami/postgres-exporter/bin:$PATH" \
    POSTGRES_PASSWORD="password" \
    POSTGRES_USER="username" \
    DATA_SOURCE_NAME="postgresql://username:password@localhost:5432/postgres?sslmode=disable"

EXPOSE 9187 5432

#WORKDIR /usr/local/bin
#COPY entrypoint.sh /opt/bitnami/postgres-exporter

WORKDIR /opt/bitnami/postgres-exporter
USER 999


#ENTRYPOINT [ "postgres_exporter" ]
#ENTRYPOINT [ "docker-entrypoint.sh" ]
ENTRYPOINT [ "/opt/bitnami/postgres-exporter/entrypoint.sh" ]