# MultiSelect Component

The `MultiSelect` component is a flexible, feature-rich dropdown designed for selecting multiple items. It supports grouping, searching, and a special "Select All" mode with exclusive logic.

## Features

- **Multi-Selection**: Select multiple items from a list.
- **Grouping**: Organize options into labeled groups.
- **Search**: Built-in search functionality (toggleable).
- **"All" Mode**: Logic for an exclusive "All" option (e.g., "All Namespaces").
- **Smart Selection**:
  - Clicking an item while "All" is active switches to "Focus Mode" (only that item selected).
  - Manually selecting all items switches back to "All" mode.
- **Dynamic Display**: Shows count or list of badges depending on selection size.

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `options` | `Option[]` | `[]` | Flat list of options to display. |
| `groups` | `Group[]` | `[]` | Grouped options. Overrides `options` if provided. |
| `selected` | `string[]` | Required | Array of selected values. |
| `onChange` | `(values: string[]) => void` | Required | Callback when selection changes. |
| `placeholder` | `string` | `"Select..."` | Placeholder text when empty. |
| `loading` | `boolean` | `false` | Loading state. |
| `allOption` | `Option` | `undefined` | Special option for "Select All" (exclusive behavior). |
| `showSearch` | `boolean` | `true` | Whether to show the search input. |

### Types

```typescript
interface Option {
    label: string;
    value: string;
    icon?: React.ComponentType<{ className?: string }>;
}

interface Group {
    label: string;
    options: Option[];
}
```

## Usage Examples

### Basic Usage

```tsx
<MultiSelect
    options={[
        { label: "Option 1", value: "opt1" },
        { label: "Option 2", value: "opt2" },
    ]}
    selected={selectedValues}
    onChange={setSelectedValues}
/>
```

### With Grouping

```tsx
<MultiSelect
    groups={[
        {
            label: "Group A",
            options: [
                { label: "Item A1", value: "a1" },
                { label: "Item A2", value: "a2" },
            ]
        },
        {
            label: "Group B",
            options: [
                { label: "Item B1", value: "b1" },
            ]
        }
    ]}
    selected={selectedValues}
    onChange={setSelectedValues}
/>
```

### With "Select All" Option

When `allOption` is provided, the component handles mutual exclusivity: identifying when "All" is meant vs. specific selections.

```tsx
<MultiSelect
    options={namespaces}
    selected={currentNamespaces}
    onChange={updateNamespaces}
    allOption={{ label: "All Namespaces", value: "__all__" }}
/>
```

If `selected` contains `__all__`, the component visually checks all items. Clicking a specific item will switch to focusing only on that item.
