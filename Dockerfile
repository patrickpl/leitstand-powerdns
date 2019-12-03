FROM alpine

EXPOSE 19991

COPY /bin/linux_amd64/connector /bin/
COPY /config/powerdns_prod.json /etc/leitstand/connector/powerdns.json


CMD /bin/connector
