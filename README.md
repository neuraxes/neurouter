# Neurouter

Neurouter is a powerful LLM router that provides a unified interface for multiple Large Language Model providers. It acts as a proxy service that can route requests to various AI model providers while presenting a consistent API interface to clients.

## Features

- 🔄 **Unified API Interface**: Single consistent API for multiple LLM providers
- 🌐 **Multiple Provider Support**:
  - OpenAI
  - Anthropic
  - Google (Gemini)
  - DeepSeek
- 🔌 **Flexible Protocol Support**:
  - gRPC (port 9000)
  - HTTP/REST (port 8000)
- ⚡ **High Performance**: Efficient request routing and handling
- 🛠 **Configurable**: Easy configuration through YAML files

## Installation

### Prerequisites

- Go 1.24 or later
- Docker (optional)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/neuraxes/neurouter.git
cd neurouter

# Build the binary
make build
```

### Using Docker

#### Using Prebuilt Container

```bash
# Run container with your configuration
# Note: Server configuration is included in the container by default
docker run -d \
  --name neurouter \
  -p 8000:8000 \
  -p 9000:9000 \
  -v $(pwd)/configs/upstream.yaml:/configs/upstream.yaml \
  ghcr.io/neuraxes/neurouter:latest
```

#### Building from Dockerfile

```bash
# Build image locally
docker build -t neurouter .

# Run container
docker run -d \
  --name neurouter \
  -p 8000:8000 \
  -p 9000:9000 \
  -v $(pwd)/configs/upstream.yaml:/configs/upstream.yaml \
  neurouter
```

## Configuration

Neurouter uses two main configuration files:

### Server Configuration (`configs/config.yaml`)

```yaml
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 120s
  grpc:
    addr: 0.0.0.0:9000
    timeout: 120s
```

### Provider Configuration (`configs/upstream.yaml`)

The upstream configuration defines available models and their properties:

```yaml
upstream:
  configs:
    - name: "provider-name"
      models:
        - id: "model-id"          # Unique identifier for the model
          upstream_id: "model-id"  # Model ID in the upstream service
          name: "Model Name"       # Display name
          from: "provider"         # The owner of model
          provider: "provider"     # The model service provider
          modalities:             # Supported modalities
            - "MODALITY_TEXT"
            - "MODALITY_IMAGE"
          capabilities:           # Model capabilities
            - "CAPABILITY_CHAT"
            - "CAPABILITY_COMPLETION"
      providerSpecific:
        ...
```

See: [Upstream Providers](#upstream-providers) for detailed provider-specific configurations.

## Usage

### HTTP API

Access the OpenAI-compatible REST API:

```bash
curl -X POST http://localhost:8000/openai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### gRPC

Use the gRPC API with your preferred gRPC client. Protocol buffer definitions can be found in `api/neurouter/v1/`.

## Upstream Providers

Neurouter supports multiple LLM providers through a unified configuration system. Each provider can be configured in `configs/upstream.yaml`.

1. **OpenAI**
   ```yaml
   name: "openai-config"
   models:
     - id: "gpt-4"
       upstream_id: "gpt-4"
       name: "GPT-4"
       from: "openai"
       provider: "openai"
       modalities: ["MODALITY_TEXT"]
       capabilities: ["CAPABILITY_CHAT", "CAPABILITY_COMPLETION"]
   open_ai:
     api_key: "your-api-key"
     base_url: "https://api.openai.com/v1"  # Optional
     prefer_string_content_for_system: false # Optional
     prefer_string_content_for_user: false   # Optional
     prefer_string_content_for_assistant: false # Optional
     prefer_string_content_for_tool: false   # Optional
     prefer_single_part_content: false       # Optional
   ```

2. **Anthropic**
   ```yaml
   name: "anthropic-config"
   models:
     - id: "claude-3"
       upstream_id: "claude-3-sonnet"
       name: "Claude 3 Sonnet"
       from: "anthropic"
       provider: "anthropic"
       modalities: ["MODALITY_TEXT"]
       capabilities: ["CAPABILITY_CHAT"]
   anthropic:
     api_key: "your-api-key"
     base_url: "https://api.anthropic.com"  # Optional
     merge_system: false                    # Whether to put system prompts into messages
   ```

3. **Google (Gemini)**
   ```yaml
   name: "google-config"
   models:
     - id: "gemini-pro"
       upstream_id: "gemini-pro"
       name: "Gemini Pro"
       from: "google"
       provider: "google"
       modalities: ["MODALITY_TEXT"]
       capabilities: ["CAPABILITY_CHAT"]
   google:
     api_key: "your-api-key"
   ```

4. **DeepSeek**
   ```yaml
   name: "deepseek-config"
   models:
     - id: "deepseek-v3"
       upstream_id: "deepseek-chat"
       name: "DeepSeek Chat"
       from: "deepseek"
       provider: "deepseek"
       modalities: ["MODALITY_TEXT"]
       capabilities: ["CAPABILITY_CHAT"]
   deep_seek:
     api_key: "your-api-key"
     base_url: "https://api.deepseek.com"
   ```

5. **Neurouter** (for chaining multiple Neurouter instances)
   ```yaml
   name: "neurouter-config"
   models:
     - id: "remote-model"
       upstream_id: "remote-model"
       name: "Remote Model"
       from: "neurouter"
       provider: "neurouter"
       modalities: ["MODALITY_TEXT"]
       capabilities: ["CAPABILITY_CHAT"]
   neurouter:
     endpoint: "another-neurouter-instance:9000"
   ```

## License

Licensed under the Apache License, Version 2.0

You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
