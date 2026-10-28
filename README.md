# hndelegate

A decision-tree based Hacker News curator built with Jev. hndelegate polls new Hacker News stories, evaluates the content against customised decision trees, and sends selected posts to a Telegram channel.

## How it works

Every two minutes, hndelegate:
- polls Hacker News for new stories.
- skips already processed stories, and fetches the content for the rest.
- evaluates them against one or more independent decision trees using the [Jev](https://typesafe.ai/blog/introducing-system-one-models-and-jev) classification model. The decision trees are fully customisable and are stored in a SQLite database. See [Decisions](docs/decisions.md) for more info.
- sends stories that successfully match at least one of the trees to the configured Telegram chat or channel.

See [Architecture](docs/architecture.md) for a detailed rundown of how hndelegate works under the hood.

## Running hndelegate

You can get the latest release [here](https://github.com/uint4096/hndelegate/releases/latest). hndelegate uses the following environment variables.

| Environment Variable      | Description                                                                                                                                 |  Status |
|---------------------------|---------------------------------------------------------------------------------------------------------------------------------------------|--------------|
| `JEV_API_KEY`             | The API key for Typesafe's Jev model. You can get one from the [Typesafe console](https://console.typesafe.ai/home).                        |  Required    |
| `TELEGRAM_BOT_TOKEN`      | The Telegram bot token. Create one with [@BotFather](https://telegram.me/botfather).                                                        |  Required    |
| `TELEGRAM_CHAT_ID`        | ID of the chat or channel to post stories to. Use [@get_id_bot](https://t.me/get_id_bot) to get the chat ID for your chat or channel.      |  Required    |
| `TELEGRAM_ERROR_CHAT_ID`  | An optional chat ID to send alerts for hndelegate errors.                                                                                   |  Optional    |
| `MAX_ITEMS_PER_ROUND`     | Maximum number of posts processed per round (every two minutes). Skipped posts are dropped completely. Defaults to 20.                      |  Optional    |
| `DB_PATH`                 | Path to the directory containing the SQLite database (creates the db if none exists). Defaults to the current directory.                    |  Optional    |

You must add at least one decision tree to the database for hndelegate to run. See [Decisions](docs/decisions.md).

### Building from source

You need the following to build hndelegate from source:

- Go 1.27+
- C compiler (required to compile go-sqlite3)

Clone this repo and run:
```shell
go build -o <path_to_output>
```

Make sure all of the required environment variables are populated in your shell before running the resulting binary.

### Running with Docker

hndelegate comes with a pre-built Docker image. To get the bot running with Docker, run:

```shell
docker pull uint4096/hndelegate:latest
docker run --env-file <path_to_env_file> uint4096/hndelegate:latest
```

To persist the SQLite database between runs, mount a volume when running `docker run`:

```shell
docker run --env-file <path_to_env_file> -v /home_dir/hndelegate:/root/.config/hndelegate:rw uint4096/hndelegate:latest
```
