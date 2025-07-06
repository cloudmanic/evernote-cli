# evernote-cli

A CLI tool for interacting with Evernote. Currently supports authentication and searching notes.

## Authentication

Before using other commands you must authenticate. Set the environment variables `EVERNOTE_CLIENT_ID` and `EVERNOTE_CLIENT_SECRET` with your developer credentials and run:

```bash
evernote-cli auth
```

This will open a browser window to authenticate with Evernote. When finished, an access token is stored in `~/.config/evernote/auth.json`.

## Searching

Search notes with:

```bash
evernote-cli search "your query"
```

Use `--json` to output the raw JSON returned by the API.

