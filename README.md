# mattermost-poll-bot

# Build

### 1. Fetch git submodules:

In the root folder run this command:

```bash
git submodule update --init --recursive
```

### 2. Run mattermost with docker

```bash
cd docker
docker compose -f docker-compose.yml -f docker-compose.without-nginx.yml up -d
```

Mattermost url: http://127.0.0.1:8065/

To shutdown deployment run this command (in `docker` folder):

```bash
docker compose -f docker-compose.yml -f docker-compose.without-nginx.yml down
```
