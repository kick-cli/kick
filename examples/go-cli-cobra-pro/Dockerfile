# syntax=docker/dockerfile:1
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY {{.project_name}} /usr/local/bin/{{.project_name}}
ENTRYPOINT ["/usr/local/bin/{{.project_name}}"]
