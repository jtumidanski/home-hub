# Weather Forecast — UX Flow

## 1. Dashboard Weather Widget

### State: Location Set

```
┌─────────────────────────────┐
│  ☀ 72°F  Partly Cloudy     │
│  H: 78°  L: 55°            │
│                             │
│  New York, NY               │
│  Updated 2:30 PM    →    │
└─────────────────────────────┘
```

- Current temperature displayed prominently
- Weather condition icon (sun, cloud, rain, snow, etc.) alongside summary text
- Today's high and low temperatures
- Location name shown for context
- Absolute timestamp of last fetch (e.g., "Updated 2:30 PM")
- Entire card is tappable/clickable, navigates to `/weather`

### State: No Location Set

```
┌─────────────────────────────┐
│  No location set            │
│                             │
│  Set your household location│
│  to see weather.            │
│                             │
│  [Go to Settings]           │
└─────────────────────────────┘
```

- Clear call-to-action linking to household settings

### State: Loading

- Skeleton/shimmer placeholder matching the widget dimensions

### State: Error (upstream unavailable, stale cache)

```
┌─────────────────────────────┐
│  ☀ 72°F  Partly Cloudy     │
│  H: 78°  L: 55°            │
│                             │
│  New York, NY               │
│  Updated 12:30 PM     ⚠    │
└─────────────────────────────┘
```

- Stale data is shown with a warning indicator and an older absolute timestamp
- If no cache exists at all, show a brief error message

---

## 2. Weather Page (`/weather`)

### State: Location Set

```
Weather — New York, NY
─────────────────────────────────

Today                  ☀ Partly Cloudy
                       H: 78°F   L: 55°F
                       ← visually highlighted

Wed Mar 26             🌧 Rain
                       H: 65°F   L: 48°F

Thu Mar 27             ☁ Overcast
                       H: 60°F   L: 45°F

Fri Mar 28             ☀ Clear
                       H: 70°F   L: 50°F

Sat Mar 29             ☀ Mostly Clear
                       H: 72°F   L: 52°F

Sun Mar 30             🌧 Rain Showers
                       H: 58°F   L: 44°F

Mon Mar 31             ☁ Partly Cloudy
                       H: 62°F   L: 46°F

─────────────────────────────────
Updated 2:30 PM
```

- Page title includes location name
- 7 cards/rows, one per day
- Today's entry is visually distinguished (highlighted background or border)
- Each day shows: day name, date, weather icon, condition summary, high/low temps
- Last updated timestamp at the bottom

### State: No Location

- Message explaining no location is set
- Link to household settings

### Mobile Layout

- Cards stack vertically, full-width
- Each card is a self-contained row (no horizontal scrolling)
- Tap-friendly sizing per user's mobile UI preferences

---

## 3. Household Settings — Location Input

### Current State: No Location

```
Location
┌─────────────────────────────┐
│  Search for a city...       │
└─────────────────────────────┘
```

### Typing / Autocomplete Active

```
Location
┌─────────────────────────────┐
│  New Yo                     │
├─────────────────────────────┤
│  New York, New York, US     │
│  New York Mills, Minnesota  │
│  New Yorktown, Indiana      │
│  ...                        │
└─────────────────────────────┘
```

- Autocomplete dropdown appears after 2+ characters
- Results show: name, admin region, country
- Debounced input (300ms) to avoid excessive API calls
- Keyboard navigable (up/down/enter)

### After Selection

```
Location
┌─────────────────────────────┐
│  New York, New York, US   ✕ │
└─────────────────────────────┘
```

- Selected place shown as a chip/tag
- Clear button (✕) to remove the location
- Latitude/longitude populated behind the scenes
- Saving the form PATCHes the household with lat, lon, and locationName

---

## 4. Navigation

- Dashboard → Weather widget → `/weather` page
- Sidebar: "Weather" item under a logical grouping (below Tasks/Reminders)
- Weather page → "Change location" link → Household settings
- Household settings includes the location field in the existing form

---

## 5. Weather Icons

The backend returns an `icon` key in weather API responses (e.g., `"cloud-rain"`). The frontend maps this key to the corresponding Lucide React component. No WMO code interpretation is needed on the frontend.

| Icon Key         | Lucide Component |
|------------------|------------------|
| sun              | Sun              |
| cloud-sun        | CloudSun         |
| cloud            | Cloud            |
| cloud-fog        | CloudFog         |
| cloud-drizzle    | CloudDrizzle     |
| cloud-rain       | CloudRain        |
| snowflake        | Snowflake        |
| cloud-lightning  | CloudLightning   |

The frontend maintains a simple lookup from icon key to component. The full WMO code-to-icon mapping lives in the backend only (see api-contracts.md).
