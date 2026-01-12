# Log Viewer Behavior Scenarios

This document outlines the behavior of the Cloud Sentinel Log Viewer across different container configurations and user interactions.

## 1. Single-Container Pods

**Scenario:** A user opens logs for a pod that contains only one container.

*   **Default View:**
    *   Logs are streamed immediately.
    *   **Prefixes:** Hidden by default. The user sees raw log output.
*   **Container Selector:**
    *   Shows: "All Containers" (Active) and the single container name.
    *   Since there is only one container, switching selection effectively does nothing to the stream content, but is available for consistency.
*   **Prefix Toggle:**
    *   **Visible:** Yes.
    *   **State:** Defaults to `OFF`.
    *   **Interaction:** User can click the tag icon to manually show `[container-name]` prefix if desired.

## 2. Multi-Container Pods

**Scenario:** A user opens logs for a pod with multiple containers (e.g., `app-container` and `sidecar-proxy`).

*   **Default View:**
    *   Logs are streamed from **ALL** containers concurrently.
    *   **Prefixes:** Visible by default (e.g., `[app-container] Log message...`).
*   **Container Selector:**
    *   **Default:** "All Containers" is selected.
    *   **Dropdown:** Lists "All Containers" (Active) and individual checkboxes for each container.
*   **Interaction Scenarios:**
    *   **Click a specific container (Focus Mode):**
*   **Default View**: "All Containers" selected. Logs from all containers stream in parallel.
*   **Prefixes**: **On** by default.
    *   **Behavior**: Prefix state is set initially. Changing selection **does not** automatically toggle prefixes (User preference persists).
    *   **Format**: `[container-name] log line...`
*   **Selection Logic**:
    *   **Mixed Selection**: User can select any subset of containers.
    *   **Focus Mode**: Clicking a specific container when "All" is active instantly selects ONLY that container.
    *   **Deselection**: User can deselect all containers (showing "Select at least one container" empty state).
    *   **Auto-Revert**: Selecting all individual containers automatically reverts to the "All Containers" sentinel state.

## 3. Init Containers
*   **Visibility**: Listed in a separate "INIT CONTAINERS" section in the dropdown.
*   **Behavior**: Fully selectable, included in "All Containers" stream.
*   **Styling**: Distinct headers in dropdown to separate from standard containers.

## 4. Technical Implementation details
*   **Concurrency**: Backend uses goroutines to stream multiple containers.
*   **Stability**: WebSocket heartbeat (Ping every 15s) prevents connection timeouts.
*   **Streaming**: Unified logic (`streamContainers`) handles both single and multi-container requests efficiently.
*   **UI**:
    *   **Dropdown**: Rounded-lg container, rounded-sm checkboxes.
    *   **Empty State**: Overlay message when no selection.
