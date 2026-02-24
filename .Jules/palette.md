## 2024-05-23 - [Settings Button Accessibility]
**Learning:** Icon-only buttons (like `Settings2` in `ResourceTable`) MUST have accessible names and visual tooltips for clarity.
**Action:** Always wrap icon buttons in `Tooltip` and provide `aria-label`. Use translation keys for consistency.

## 2024-05-23 - [Frontend Redirect Logic]
**Learning:** The frontend forces redirect to `/settings` if no clusters are configured for admin users, blocking access to other routes like `/pods`.
**Action:** When testing components on protected routes, either mock the cluster context or temporarily disable the redirect logic.

## 2026-02-24 - [Sidebar Customizer Accessibility]
**Learning:** Many icon-only buttons in the `SidebarCustomizer` relied solely on `title` attributes, making them inaccessible to screen readers.
**Action:** Ensure all icon-only buttons have an `aria-label` that mirrors the `title` or provides a descriptive name using translation keys where possible.

## 2026-02-24 - [Conventional Commits]
**Learning:** PR titles must strictly follow Conventional Commits format (e.g., `feat: ...`, `fix: ...`) to pass CI checks. Creative titles like "🎨 Palette: ..." cause validation failures.
**Action:** Always start PR titles with a standard type prefix. Put creative flair in the description or body.
