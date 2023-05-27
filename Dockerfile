FROM golang AS build-stage

# Set destination for COPY
WORKDIR /app

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY go.mod go.sum *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /pickmypostcode

FROM ghcr.io/go-rod/rod

WORKDIR /

COPY --from=build-stage /pickmypostcode /pickmypostcode

EXPOSE 1234

# Run
CMD ["/pickmypostcode"]