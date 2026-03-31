# Ingredient Categories — UX Flow

## 1. Category Management

Accessible from the Ingredients page (new tab or section).

### Category List View
- Table showing all categories: name, ingredient count, sort order
- Default categories are auto-seeded on first visit (the list endpoint handles this transparently)
- Inline add: text input + "Add" button at bottom of list
- Inline rename: click category name to edit in place
- Delete: icon button with confirmation dialog ("X ingredients will become uncategorized")
- Sort order is displayed but not user-editable (hardcoded aisle order for now)

## 2. Bulk Category Assignment

Accessible from the Ingredients page (new mode or dedicated view).

### Bulk Edit Flow
1. User enters bulk-edit mode (toggle button on Ingredients page)
2. Filter controls appear:
   - Dropdown: "All" / "Uncategorized" / specific category
   - Search: existing name search
   - Badge showing count of uncategorized ingredients
3. Ingredient list shows checkboxes for multi-select
4. "Select All" checkbox in header (selects visible/filtered ingredients)
5. Sticky action bar appears when 1+ ingredients selected:
   - Shows selection count
   - Category dropdown selector
   - "Apply" button
6. On apply: calls bulk-categorize endpoint, refreshes list, clears selection

### Single Ingredient Category Edit
- Ingredient detail page shows a category dropdown field alongside existing fields (name, display name, unit family)
- Dropdown lists all categories + "Uncategorized" option

## 3. Ingredient Preview (Plan View)

The consolidated ingredient list in the meal plan view changes from flat to grouped.

### Grouped Display
```
Produce
  2 large  tomato
  500 g    spinach
  3        onion

Dairy & Eggs
  250 ml   whole milk
  200 g    cheddar cheese

Uncategorized
  1 cup    mystery ingredient (italic, muted)
```

- Category headers are styled as section dividers (bold text, subtle top border)
- Categories appear in sort_order sequence
- "Uncategorized" always appears last
- Empty categories are omitted
- Within each category, ingredients sorted alphabetically by display name
- Unresolved ingredients retain their existing italic/muted styling

## 4. Markdown Export

The markdown export groups ingredients under category headers:

```markdown
## Shopping List

### Produce
- 2 large tomato
- 500 g spinach
- 3 onion

### Dairy & Eggs
- 250 ml whole milk
- 200 g cheddar cheese

### Uncategorized
- 1 cup mystery ingredient _(unresolved)_
```

- Categories in sort_order sequence
- "Uncategorized" last
- Empty categories omitted
