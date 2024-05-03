FROM python:3.12-slim

RUN apt-get update && \
    apt-get install --no-install-recommends -y git

# Set the working directory in the container
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

# Install any needed dependencies specified in requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Make sure all messages always reach console
ENV PYTHONUNBUFFERED=1

USER 65534

# Run scanner when the container launches
CMD ["python3", "/app/scanner.py"]
