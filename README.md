# Git Snap

A generic framework for correlating timestamped events with git commits based on configurable attribute matching and temporal proximity.

## Overview

Git Snap enables automated tagging and analysis of commits based on external system events like AI inference calls, deployments, build events, and more. It provides a flexible correlation engine that can "snap" any event to relevant commits using configurable rules.

## Features

- **Generic Event Correlation**: Correlate any timestamped event with git commits
- **Configurable Matching**: Support for exact, contains, regex, and fuzzy matching
- **Multiple Input Formats**: JSON, JSONL, and CSV event file support
- **Temporal Scoring**: Weight correlations based on time proximity
- **Flexible Configuration**: YAML-based configuration with predefined templates
- **GitHub Integration**: Enhanced commit data from GitHub API
- **CLI Interface**: Easy-to-use command-line tool

## Installation

```bash
go install github.com/fraser-isbester/git-snap/cmd/git-snap@latest
```

Or build from source:

```bash
git clone https://github.com/fraser-isbester/git-snap.git
cd git-snap
go build -o bin/git-snap ./cmd/git-snap
```

## Quick Start

1. Initialize git-snap in your repository:
   ```bash
   git-snap init
   ```

2. Prepare your events file (see [examples](examples/)):
   ```json
   [
     {
       "id": "inference_001",
       "timestamp": "2024-01-15T10:30:00Z",
       "attributes": {
         "user_id": "john.doe@company.com",
         "project": "backend-service"
       }
     }
   ]
   ```

3. Run correlation:
   ```bash
   git-snap correlate -e events.json -c ai-inference -o table
   ```

## Usage

### Basic Correlation

```bash
# Correlate events with commits using default configuration
git-snap correlate -e events.json

# Use specific configuration
git-snap correlate -e events.json -c ai-inference

# Set minimum score threshold
git-snap correlate -e events.json -t 0.7

# Output as table
git-snap correlate -e events.json -o table
```

### Configuration Management

```bash
# List available configurations
git-snap config list

# Show configuration details
git-snap config show ai-inference

# Initialize default configurations
git-snap config init
```

### Supported Event Formats

#### JSON
```json
[
  {
    "id": "event1",
    "timestamp": "2024-01-15T10:30:00Z",
    "attributes": {
      "user_id": "john.doe@company.com",
      "project": "backend-service"
    }
  }
]
```

#### JSONL
```jsonl
{"id": "event1", "timestamp": "2024-01-15T10:30:00Z", "attributes": {"user_id": "john.doe@company.com"}}
{"id": "event2", "timestamp": "2024-01-15T10:45:00Z", "attributes": {"user_id": "jane.smith@company.com"}}
```

#### CSV
```csv
id,timestamp,user_id,project
event1,2024-01-15T10:30:00Z,john.doe@company.com,backend-service
event2,2024-01-15T10:45:00Z,jane.smith@company.com,frontend-app
```

## Configuration

Git Snap uses YAML configuration files to define correlation rules:

```yaml
time_window: "15m"
attribute_rules:
  - event_key: "user_id"
    commit_key: "author_email"
    match_type: "exact"
    required: true
  - event_key: "project"
    commit_key: "repository"
    match_type: "contains"
    required: false
score_weights:
  temporal: 0.6
  attribute: 0.4
```

### Built-in Configurations

- **default**: Basic correlation with user matching
- **ai-inference**: Optimized for AI inference events
- **deployment**: Designed for deployment correlation

## Architecture

### Core Components

1. **Event Parser**: Handles JSON, JSONL, and CSV formats
2. **Git Client**: Extracts commit data from repositories
3. **Correlation Engine**: Matches events to commits using configurable rules
4. **Configuration Manager**: Manages YAML-based correlation configurations
5. **CLI Interface**: Command-line tool for easy interaction

### Scoring Algorithm

```
Total Score = (Temporal Score × Temporal Weight) + (Attribute Score × Attribute Weight)

Temporal Score = 1.0 - (TimeDelta / TimeWindow)
Attribute Score = (Matched Required + Matched Optional) / (Total Required + Total Optional)
```

## Examples

See the [examples](examples/) directory for sample event files and usage patterns.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Use Cases

- **AI-Assisted Development**: Loosely correlate which commits were created with AI assistance
- **Deployment Correlation**: Link deployments to specific commits
- **Build Analysis**: Correlate build events with code changes
- **Developer Workflow**: Analyze patterns in development activities
- **Compliance Auditing**: Maintain trails for regulatory requirements
