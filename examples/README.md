# Git Snap Examples

This directory contains example event files to demonstrate git-snap functionality.

## Example Files

### AI Inference Events (`ai-inference-events.json`)
Contains AI inference API calls with user information and project context. Use with the `ai-inference` configuration:

```bash
git-snap correlate -e examples/ai-inference-events.json -c ai-inference
```

### Deployment Events (`deployment-events.csv`)
Contains deployment events with commit SHAs and environment information. Use with the `deployment` configuration:

```bash
git-snap correlate -e examples/deployment-events.csv -c deployment
```

### Build Events (`build-events.jsonl`)
Contains CI/CD build events with commit information. Use with the `default` configuration:

```bash
git-snap correlate -e examples/build-events.jsonl -c default
```

## Quick Start

1. Initialize git-snap:
   ```bash
   git-snap init
   ```

2. Run correlation with an example:
   ```bash
   git-snap correlate -e examples/ai-inference-events.json -c ai-inference -o table
   ```

3. View available configurations:
   ```bash
   git-snap config list
   ```

4. Show configuration details:
   ```bash
   git-snap config show ai-inference
   ```

## Configuration Examples

### AI Inference Configuration
- **Time Window**: 15 minutes
- **Required Match**: user_id -> author_email (exact)
- **Optional Match**: project -> repository (contains)
- **Scoring**: 60% temporal, 40% attribute

### Deployment Configuration
- **Time Window**: 2 hours
- **Required Match**: commit_sha -> sha (exact)
- **Optional Match**: environment -> branch (regex)
- **Scoring**: 30% temporal, 70% attribute

### Default Configuration
- **Time Window**: 15 minutes
- **Required Match**: user_id -> author_email (exact)
- **Scoring**: 50% temporal, 50% attribute./bin/git-snap config show default