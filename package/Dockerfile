#
# build the target application
#
FROM public.ecr.aws/docker/library/golang:1.20-alpine3.17 as builder
RUN apk update && apk upgrade && apk add --no-cache make yarn

# see: https://stackoverflow.com/questions/69692842/error-message-error0308010cdigital-envelope-routinesunsupported
ENV NODE_OPTIONS --openssl-legacy-provider

WORKDIR /build
COPY go.mod go.sum Makefile ./
COPY backend ./backend
COPY frontend ./frontend
RUN make linux-srv web

#
# build the target container
#
FROM public.ecr.aws/docker/library/alpine:3.17

# update the packages
RUN apk update && apk upgrade && apk add bash tzdata ca-certificates curl && rm -fr /var/cache/apk/*

# Create the run user and group
RUN addgroup --gid 18570 sse && adduser --uid 1984 docker -G sse -D

# set the timezone appropriatly
ENV TZ=UTC
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Specify home 
ENV APP_HOME /apollo
WORKDIR $APP_HOME

# Create necessary directories
RUN mkdir -p $APP_HOME/scripts $APP_HOME/bin $APP_HOME/db
RUN chown -R docker $APP_HOME && chgrp -R sse $APP_HOME

# add the migrate tool
RUN cd $APP_HOME/bin && curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz

# port and run command
EXPOSE 8080
CMD scripts/entry.sh

# Move in necessary assets
COPY package/data/container_bash_profile /home/docker/.profile
COPY package/scripts/entry.sh $APP_HOME/scripts/entry.sh
COPY backend/db/migrations/*.sql $APP_HOME/db/
COPY templates $APP_HOME/bin/templates
COPY --from=builder /build/bin/apollosvr.linux $APP_HOME/bin/apollo
COPY --from=builder /build/bin/public $APP_HOME/bin/public

# Ensure permissions are correct
RUN chown docker:sse /home/docker/.profile $APP_HOME/scripts/entry.sh $APP_HOME/bin/apollo $APP_HOME/bin/migrate $APP_HOME/db && chmod 755 /home/docker/.profile $APP_HOME/scripts/entry.sh $APP_HOME/bin/apollo $APP_HOME/bin/migrate

# Add the build tag
ARG BUILD_TAG
RUN test -n "$BUILD_TAG" && touch $APP_HOME/buildtag.build-$BUILD_TAG || touch $APP_HOME/buildtag.build-0

# Specify the user
USER docker

#
# end of file
#
