# mattermost-poll-bot

## Build

#### 1. Fetch git submodules:

In the root folder run this command:

```bash
git submodule update --init --recursive
```

#### 2. Run mattermost with docker

```bash
cd docker
```

Create config:
```bash
cp env.example .env
```

Edit `.env`: change `DOMAIN` field value to: 127.0.0.1

Then create required directories and set permissions:

```bash
mkdir -p ./volumes/app/mattermost/{config,data,logs,plugins,client/plugins,bleve-indexes}
sudo chown -R 2000:2000 ./volumes/app/mattermost
```

Deploy mattermost:

```bash
docker compose -f docker-compose.yml -f docker-compose.without-nginx.yml up -d
```

Mattermost url: http://127.0.0.1:8065/

To shutdown deployment run this command (in `docker` folder):

```bash
docker compose -f docker-compose.yml -f docker-compose.without-nginx.yml down
```

#### 3. Create .env file

...

#### 4. Build and run poll-bot

From root folder:

```bash
cd poll-bot/cmd/poll_bot
go mod tidy
go run .
```


