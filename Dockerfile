FROM image:tag
WORKDIR /

RUN dnf -y update

RUN dnf install -y iproute iputils bind-utils file hostname procps net-tools dnf-plugins-core findutils

ADD home-simplecert /usr/sbin/home-simplecert
RUN chmod +x /usr/sbin/home-simplecert

CMD ["/usr/sbin/home-simplecert-server", "run", "-c", "/etc/home-simplecert-server.yaml"]