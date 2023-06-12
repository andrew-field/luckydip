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

# sudo systemctl start docker
# sudo docker build -t pick .
# sudo docker run -m=2gb -p 1234:1234 pick The -p option is coupled with browser.ServeMonitor("0.0.0.0:1234") in main
# sudo docker images
# sudo docker container prune
# sudo docker image rm pick
# sudo docker image prune
