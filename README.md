# evernote-cli

A CLI tool for interacting with Evernote. Currently supports authentication and searching notes.

## Initial Setup

Run `evernote-cli init` to store your Evernote developer credentials. You will be prompted for your client ID and secret which are saved to `~/.config/evernote/auth.json`.

## Authentication

Once initialized, obtain an OAuth token with:

```bash
evernote-cli auth
```

This opens a browser window to authenticate with Evernote. The access token is saved alongside your credentials in `~/.config/evernote/auth.json`.

## Searching

Search notes with:

```bash
evernote-cli search "your query"
```

Use `--json` to output the raw JSON returned by the API.

