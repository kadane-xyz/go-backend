# Use the existing Judge0 container as the base image
FROM judge0/judge0:1.13.1

# Copy the judge0.conf configuration file into the container
COPY judge0.conf /judge0.conf

# Ensure the appropriate user has access, if required by Judge0
USER judge0