# syntax=docker/dockerfile:1
FROM node:19 as build

ARG RELEASE_NAME

# create a working directory inside the image
WORKDIR /dashboard

COPY . .

# Install deps
RUN npm ci

# Copy env file
COPY .env ./

# Build
RUN echo "Release name: $RELEASE_NAME"
RUN VITE_SENTRY_RELEASE=${RELEASE_NAME} npm run build

# This is only here so that we can export the built artifacts
FROM scratch as artifact
COPY --from=build /dashboard/build /
