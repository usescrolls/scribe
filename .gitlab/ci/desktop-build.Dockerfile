ARG GO_VERSION=1.26
FROM golang:${GO_VERSION}

ARG ZIG_VERSION=0.14.1
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    pkg-config \
    xz-utils \
    libayatana-appindicator3-dev \
    libgtk-3-dev \
    libwebkit2gtk-4.1-dev \
 && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL "https://ziglang.org/download/${ZIG_VERSION}/zig-x86_64-linux-${ZIG_VERSION}.tar.xz" \
    | tar -xJ -C /usr/local \
 && ln -s "/usr/local/zig-x86_64-linux-${ZIG_VERSION}/zig" /usr/local/bin/zig \
 && printf '#!/bin/sh\nexec zig cc -target aarch64-macos "$@"\n' > /usr/local/bin/zcc \
 && chmod +x /usr/local/bin/zcc \
 && printf '#!/bin/sh\nexec zig c++ -target aarch64-macos "$@"\n' > /usr/local/bin/zxx \
 && chmod +x /usr/local/bin/zxx
