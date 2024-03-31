docker build . pusher:latest
docker run -d --restart always --name pusher -v ${PWD}/config:/config pusher:latest
