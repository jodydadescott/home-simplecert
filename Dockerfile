FROM image:tag
WORKDIR /

RUN dnf -y update

RUN dnf install -y iproute iputils bind-utils file hostname procps net-tools dnf-plugins-core findutils

ADD home-simplecert /usr/bin/home-simplecert
RUN chmod +x /usr/bin/home-simplecert