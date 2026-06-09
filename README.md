# ai-commit-gen

Draft a [Conventional Commits](https://www.conventionalcommits.org/) message from your staged
changes using a local language model. No API key, no cloud, no telemetry — it talks to a local
[Ollama](https://ollama.com) instance by default, and to any OpenAI-compatible endpoint if you
point it at one.

```
$ git add .
$ ai-commit-gen
feat(auth): add refresh-token rotation

Rotate refresh tokens on every use and revoke the previous one to
shrink the replay window.
```

## Why

Writing good commit messages is a chore, so people don't. This reads the actual diff, summarises
the *intent*, and formats it correctly — locally, so your code never leaves the machine.

## How the AI works (no paid dependency)

The tool sends the diff to a chat-completions endpoint and cleans up the reply. By default that
endpoint is your local Ollama (`http://localhost:11434/v1`), so everything runs offline and free.
Set `OPENAI_BASE_URL` / `OPENAI_API_KEY` to use a hosted provider instead — the wire format is the
same.

## Install

```bash
# 1. Have a model available locally (one-time, ~1.3 GB)
ollama pull llama3.2:1b
ollama serve            # if it isn't already running

# 2. Build the CLI
go build -o ai-commit-gen .
# or: go install github.com/gaurav-oberoi/ai-commit-gen@latest
```

## Usage

```bash
git add <files>
ai-commit-gen                       # print a suggested message
ai-commit-gen --commit              # create the commit directly
ai-commit-gen --all                 # describe staged + unstaged changes
ai-commit-gen --model llama3.2:1b   # pick a model
ai-commit-gen --base-url http://localhost:11434/v1
```

| Flag / env | Default | Purpose |
|------------|---------|---------|
| `--model`, `OPENAI_MODEL` | `llama3.2` | model name |
| `--base-url`, `OPENAI_BASE_URL` | local Ollama | chat-completions endpoint |
| `OPENAI_API_KEY` | _(empty)_ | sent as a bearer token when set |
| `--all` | off | include unstaged changes |
| `--commit` | off | run `git commit` with the message |

## Run with Docker

```bash
docker build -t ai-commit-gen .
docker run --rm -v "$PWD:/repo" --network host \
  -e OPENAI_BASE_URL=http://localhost:11434/v1 ai-commit-gen
```

## Tests

```bash
go test ./...
```

Covers diff/parse helpers, message cleanup (code fences, quotes, truncation), and the HTTP client
against a stub server — so the suite runs without a model present.

## License

MIT
