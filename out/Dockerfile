FROM ubuntu:24.04

WORKDIR /app

COPY deepsight /app/deepsight
COPY configs /app/configs

RUN chmod +x /app/deepsight

RUN cat > /etc/apt/sources.list.d/ubuntu.sources <<'EOF'
Types: deb
URIs: http://mirrors.ustc.edu.cn/ubuntu
Suites: noble noble-updates noble-backports
Components: main restricted universe multiverse
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg

Types: deb
URIs: http://mirrors.ustc.edu.cn/ubuntu
Suites: noble-security
Components: main restricted universe multiverse
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg
EOF

RUN apt-get update && apt-get install -y ca-certificates

EXPOSE 8080

CMD ["/app/deepsight"]
