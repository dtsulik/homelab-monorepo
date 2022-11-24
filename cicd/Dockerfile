FROM alpine

RUN mkdir /app
COPY --chown=1000:1000 build/app /app/
ENTRYPOINT /app/app