# evernote-cli

A CLI tool for interacting with Evernote. Currently supports authentication, searching notes, and listing notebooks.

## Initial Setup

Run `evernote-cli init` and follow the prompts to provide your Evernote developer client ID and secret. The command then opens a browser to authenticate and stores the resulting token together with your credentials in `~/.config/evernote/auth.json`.

## Searching

Search notes with:

```bash
evernote-cli search "your query"
```

Use `--json` to output the raw JSON returned by the API.

## Listing Notebooks

List all available notebooks with:

```bash
evernote-cli notebooks
```

This will display a formatted list of notebooks with their names and GUIDs. Use `--json` to output the raw JSON returned by the API.

