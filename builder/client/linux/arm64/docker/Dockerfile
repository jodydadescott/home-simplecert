FROM arm64v8/fedora:37
WORKDIR /

RUN dnf -y update

RUN dnf install -y iproute iputils bind-utils file hostname procps net-tools dnf-plugins-core findutils

ADD home-simplecert-client /usr/sbin/home-simplecert-client
RUN chmod +x /usr/sbin/home-simplecert-client

CMD ["/usr/sbin/home-simplecert-client", "run", "-c", "/etc/home-simplecert-client.yaml"]