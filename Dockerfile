FROM alpine

COPY ./bin/cache-story /bin/cache-story

EXPOSE 8018

ENTRYPOINT [ "/bin/cache-story" ]
