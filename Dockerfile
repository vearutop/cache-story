FROM alpine

COPY ./bin/cache-story /bin/cache-story

EXPOSE 80

ENTRYPOINT [ "/bin/cache-story" ]
