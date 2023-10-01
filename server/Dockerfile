FROM alpine:3.17

ENV NODE_VERSION 20.8.0

RUN 

COPY docker-entrypoint.sh /usr/local/bin/
ENTRYPOINT ["docker-entrypoint.sh"]


EXPOSE 4000

ENV PORT 4000
# set hostname to localhost
ENV HOSTNAME "0.0.0.0"
CMD [ "node" ]