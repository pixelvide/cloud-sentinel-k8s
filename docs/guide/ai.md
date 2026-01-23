# AI Features

Cloud Sentinel K8s integrates advanced AI capabilities to help you manage and understand your Kubernetes clusters more effectively.

## AI Assistant

The platform features an integrated AI assistant that can:
- Explain complex Kubernetes resources.
- Help troubleshoot issues based on logs and resource states.
- Generate Kubernetes manifests or commands.
- Provide insights into cluster health and configuration.

## AI Provider Profiles

Administrators can configure multiple AI provider profiles (e.g., Google Gemini, OpenAI). Profiles can be:
- **System Profiles**: Configured by administrators for use across the entire platform.
- **User Profiles**: Individually configured by users with their own API keys.

### System Governance

Cloud Sentinel K8s provides granular control over how AI services are used:
- **Allow User Keys**: When enabled, users can add their own API keys to use specific models or providers.
- **Force User Keys**: Requires users to provide their own keys; system-wide credentials will not be used.
- **Allowed Models**: Administrators can restrict which LLM models are available for each provider profile.

## Model Context Protocol (MCP)

Cloud Sentinel K8s supports the **Model Context Protocol (MCP)**, allowing the AI to safely interact with your cluster and external tools. This enables the assistant to:
- Directly read resource specifications.
- Analyze live logs.
- Execute safe diagnostic commands (if permitted).

## Configuration

To get started with AI features:
1. Navigate to **Settings > AI Administration** (Admin only) to set up provider profiles.
2. In **Settings > AI Configuration**, users can select their preferred model and provide personal API keys if enabled by the administrator.
3. Access the AI Chat interface from the side navigation or via the floating chat icon.

## Enabling for Users

Administrators can enable or disable AI chat access for specific users in the **Settings > User Management** section.
