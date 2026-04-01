## 2026-02-13 - [Repeated Computation in Column Accessors]
**Learning:** Table column accessors in `SimpleTable` implementations often call the same expensive helper function (e.g., `getPodStatus`) multiple times per row for different columns. This results in O(Columns * Rows) complexity instead of O(Rows).
**Action:** Memoize expensive calculations at the row level (e.g., using a `Map` in `useMemo` keyed by object ID) before passing to the table, or ensure the helper function itself is memoized.

## 2026-02-13 - [Memoization of Helper Functions]
**Learning:** When multiple components or table columns call the same expensive helper function (like `getPodStatus`) with the same object reference, memoizing the helper function itself using a `WeakMap` is a clean and effective optimization that avoids refactoring all call sites.
**Action:** Use `WeakMap` to cache results of expensive stateless functions that take an object as input.

## 2026-02-13 - [Stable Keys in Generic Tables]
**Learning:** Generic table components (like `SimpleTable`) were using array indices as React keys (`key={rowIndex}`). This forces unnecessary re-renders of all rows when the list order changes (e.g. sorting) and can lead to incorrect state preservation.
**Action:** Add an optional `getRowId` prop to generic table components and use stable IDs (like `metadata.uid`) as keys whenever possible.

## 2026-02-17 - [Initial Bundle Size Reduction]
**Learning:** The React application was loading the entire bundle upfront (~1.9MB), causing slow initial load times for users even if they only needed a single page. Static imports in `routes.tsx` were the primary cause.
**Action:** Use `React.lazy` and `Suspense` for route-based code splitting to significantly reduce the initial bundle size (achieved ~50% reduction).
