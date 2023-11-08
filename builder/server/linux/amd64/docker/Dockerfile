FROM fedora:37
WORKDIR /

RUN dnf -y update

RUN dnf install -y iproute iputils bind-utils file hostname procps net-tools dnf-plugins-core findutils

ADD home-simplecert-server /usr/sbin/home-simplecert-server
RUN chmod +x /usr/sbin/home-simplecert-server

CMD ["/usr/sbin/home-simplecert-server", "run", "-c", "/etc/home-simplecert-server.yaml"]