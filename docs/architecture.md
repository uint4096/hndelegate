# Architecture

hndelegate is a single Go binary that runs in a loop. It fetches new stories from Hacker News every two minutes, evaluates them against your decision trees, and sends the accepted posts to Telegram. It uses a SQLite database to manage state.

![Architecture](https://github.com/user-attachments/assets/e69ea174-257e-4640-98fe-3fb8bb2b8b2f)

## The loop

Each round runs the following steps:

1. **Construct decision tree.** Reads the decisions and choices from SQLite and builds the trees in memory. This happens every round, so changes to the trees take effect without a restart.
2. **Fetch story ids.** Fetches the latest story IDs from the HN API and removes the ones it has already processed. If more than `MAX_ITEMS_PER_ROUND` stories are left, it keeps the newest ones and drops the rest.
3. **Process stories.** Stories are processed in their respective goroutines, with at most five running at a time. See [Processing a story](#processing-a-story).
4. **Record processed posts.** Once all goroutines finish, every story from the round is written to SQLite in a single batch. This allows the next round to deduplicate already processed stories.
5. **Sleep.** Sleeps for two minutes and starts over.

## Processing a story

Each goroutine runs the following steps for a single story.

1. **Fetch HN item.** Fetches the story details from the HN API. It skips anything that isn't a story, has no URL, is a Show HN post, or has a title that already exists in the database.
2. **Fetch content.** Downloads the linked page and extracts the readable text with [go-readability](https://codeberg.org/readeck/go-readability). Only HTML pages are supported, and the text is cut off at 3000 words.
3. **Model call.** Sends the text to Jev along with every decision from every tree in a single request.
4. **Accepted.** Walks the trees using Jev's answers. A story is accepted if every decision on at least one path from a root to a leaf passes. Accepted stories are sent to Telegram. See [Decisions](decisions.md) for how a decision passes.

## Failures

If a story fails at any step, the error is logged and sent as an alert to `TELEGRAM_ERROR_CHAT_ID` if it's set. The story is still recorded, and isn't retried.

If fetching story IDs fails, the round is skipped. The process exits after three consecutive failures.
