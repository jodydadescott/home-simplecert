FROM fedora:37
WORKDIR /

RUN dnf -y update

RUN dnf install -y iproute iputils bind-utils file hostname procps net-tools dnf-plugins-core findutils

COPY home-simplecert /usr/sbin/home-simplecert
RUN chmod +x /usr/sbin/home-simplecert

CMD ["home-simplecert", "run"]