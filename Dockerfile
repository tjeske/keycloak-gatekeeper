FROM ubuntu:18.04

ADD keycloak-gatekeeper /root/
ADD apps.conf /root/
ADD config.conf /root/

WORKDIR /root

ENTRYPOINT [ "/root/keycloak-gatekeeper", "--config=config.conf", "--secure-cookie=false" ]