# syntax=docker/dockerfile:1

##
## STEP 1 - BUILD
##

FROM golang:1.20-alpine AS build

# create a working directory inside the image
WORKDIR /app

RUN apk add --no-cache make

COPY . .

# Install deps
# TODO: Avoid invalidating the cache on this step
RUN make install

# Copy env, dexrc and firebase config files
COPY .env ./
COPY .dexrc.json ./
COPY .firebase.json ./

# Copy public folder and html template files
COPY public ./public
COPY resources/templates/* ./resources/templates/

# Build
RUN CGO_ENABLED=0 make build


##
## STEP 2 - DEPLOY
##
FROM scratch

WORKDIR /

COPY --from=build /app/main.out /api
COPY --from=build /app/.env /.env
COPY --from=build /app/.dexrc.json /.dexrc.json
COPY --from=build /app/.firebase.json /.firebase.json

COPY --from=build /app/public ./public
COPY --from=build /app/resources/templates/* /resources/templates/
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 8080

ENTRYPOINT ["/api", "-env=.env", "-dex=.dexrc.json", "-firebase=.firebase.json"]
