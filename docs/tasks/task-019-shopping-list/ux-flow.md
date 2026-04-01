# Shopping List — UX Flow

## Navigation

- Sidebar entry: "Shopping" (under the existing navigation structure)
- Default view: Active shopping lists

## Page Structure

### Shopping Lists Page (`/shopping`)

**Active Lists Tab (default)**
- List of cards showing: name, item count, checked count (e.g., "12 items, 3 checked"), last updated
- "New List" button → inline name input or dialog
- Clicking a list → navigates to list detail

**Archived Lists Tab**
- List of cards showing: name, item count, archived date
- Each card has a delete action
- Clicking a list → navigates to read-only list detail

### List Detail Page (`/shopping/:id`)

Two modes controlled by a toggle or button:

#### Edit Mode (default for active lists)

- List name displayed as editable header (inline edit)
- Quick-add input at top: type a name and press Enter to add immediately (quantity and category optional, can be edited after)
- Full "Add Item" form available via expand or separate row:
  - Name field (required)
  - Quantity field (freeform text, optional)
  - Category dropdown (optional, populated from category-service)
- Items grouped by category, sorted by category sort_order
  - Within each group, items sorted by position
  - Uncategorized items appear in an "Uncategorized" group at the bottom
- Each item shows: name, quantity, category badge
- Each item has: edit (inline), delete actions
- Action bar: "Import from Meal Plan" button, "Start Shopping" button
- "Import from Meal Plan" → dialog showing available meal plans → select one → items appended

#### Shopping Mode

- Triggered by "Start Shopping" button
- Clean, focused UI optimized for mobile use
- Items grouped by category (same ordering as edit mode)
- Each item shows: checkbox, name, quantity
- Tapping an item toggles its checked state
- Checked items get strikethrough styling and move to bottom of their category group (or a collapsed "In Cart" section)
- No add/edit/delete actions visible
- Progress indicator: "8 of 24 items"
- "Uncheck All" button → resets all items (for re-shopping the list)
- "Finish Shopping" button → confirms → archives the list → redirects to shopping lists page
- "Back to Edit" button → exits shopping mode without archiving

#### Archived View (for archived lists)

- Read-only display of all items with their checked/unchecked state
- Items shown in same category-grouped layout
- "Reopen List" button → unarchives, returns to active edit mode
- "Delete List" button and navigation back
- Banner indicating "Archived on {date}"

## Meal Plan Integration

On the existing Meal Plan detail page:
- Add an "Add to Shopping List" button (near the existing export button)
- Clicking opens a dialog:
  - Shows existing active shopping lists as selectable options
  - Option to create a new list (name input)
  - "Add Ingredients" button → triggers import → shows success toast with count of items added
  - Link to navigate to the target list after import
