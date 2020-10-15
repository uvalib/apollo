#
# build the target application
#
FROM golang:1.15 as builder
RUN apt-get -y update
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
RUN apt-get -y update && apt-get -y install yarn

WORKDIR /build
COPY go.mod go.sum Makefile ./
COPY backend ./backend
COPY frontend ./frontend
RUN make linux-srv web

#
# build the target container
#
FROM alpine:3.12

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
RUN mkdir -p $APP_HOME/scripts $APP_HOME/bin
RUN chown -R docker $APP_HOME && chgrp -R sse $APP_HOME

# add the migrate tool
RUN cd $APP_HOME/bin && curl -L https://github.com/golang-migrate/migrate/releases/download/v4.6.1/migrate.linux-amd64.tar.gz | tar xvz && mv migrate.linux-amd64 migrate

# port and run command
EXPOSE 8080
CMD scripts/entry.sh

# Move in necessary assets
COPY package/scripts/entry.sh $APP_HOME/scripts/entry.sh
COPY package/data/container_bash_profile /home/docker/.profile
COPY templates $APP_HOME/bin/templates
COPY --from=builder /build/bin/apollosvr.linux $APP_HOME/bin/apollo
COPY --from=builder /build/bin/public $APP_HOME/bin/public

# Ensure permissions are correct
RUN chown docker:sse /home/docker/.profile $APP_HOME/scripts/entry.sh $APP_HOME/bin/apollo && chmod 755 /home/docker/.profile $APP_HOME/scripts/entry.sh $APP_HOME/bin/apollo

# Add the build tag
ARG BUILD_TAG
RUN test -n "$BUILD_TAG" && touch $APP_HOME/buildtag.build-$BUILD_TAG || touch $APP_HOME/buildtag.build-0

# Specify the user
USER docker

#
# end of file
#
