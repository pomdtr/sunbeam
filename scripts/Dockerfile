FROM sunbeam/sunbeam:latest

# Install dependencies
RUN npm install -g zx

WORKDIR /workspaces/sunbeam

# Copy all scripts to the image
COPY . .

RUN ["sunbeam", "-serve", "."]
