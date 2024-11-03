# Start from the official redocly/redoc image
FROM redocly/redoc

# Copy the OpenAPI YAML file into the container
COPY nginx.conf /etc/nginx/
COPY openapi.yaml /usr/share/nginx/html/swagger.yaml

# Optionally expose a different port if needed (default is 80)
EXPOSE 80