# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Members of one or a few households: adults managing the shared running of a
home, with mixed technical comfort. Also anyone in the household glancing at
the wall kiosk while walking past.

## Product Purpose

Home Hub is a self-hosted household organiser. It brings meals, recipes,
ingredients, shopping lists, a wish list, calendars, tasks, reminders,
habit trackers, workouts, package tracking and weather into one place that
every member of a household shares. Success is a household that plans and
runs its week from one calm, trustworthy surface instead of a pile of
separate apps.

## Operating Context

Three real contexts, each with different demands:

1. **Phone, one-handed**, in a shop or a kitchen: shopping mode, today's
   tasks, a recipe on the counter.
2. **Desktop planning sessions**: the meal plan, the week's workouts, the
   calendar.
3. **Wall kiosk** in a browser, read from 2–3 m away and rarely touched:
   configurable dashboards of widgets.

## Capabilities and Constraints

- Responsive web app, used installed-PWA-style on phones and full-screen on
  a kiosk browser.
- Multi-tenant: data is scoped to a household and shared by its members.
- Light and dark themes, both first-class.
- Self-hosted, with no third-party runtime requests (fonts and assets ship
  with the app).

## Brand Commitments

- Name: Home Hub.
- Tone: warm, calm, competent and quietly friendly, like a well-kept
  kitchen notebook. Not cute, not corporate.
- Personality: domestic, dependable, unhurried.
- Anti-references: dense enterprise admin dashboards; neon "smart home"
  dark UIs; pastel cottagecore.

## Evidence on Hand

Real household data only. There is no marketing copy, no testimonials and
no usage statistics; future work must not invent placeholder copy,
testimonials or stats.

## Product Principles

1. One shared household surface, not a collection of tools.
2. Calm over busy: the home should feel unhurried, even when the week isn't.
3. Readable where it is used: one hand in a shop, a glance from across the
   kitchen.
4. The household's own data is the content; nothing is faked to fill space.

## Accessibility & Inclusion

WCAG 2.2 AA in both light and dark themes. Mixed technical comfort across
users means status is never conveyed by colour alone and controls stay
large enough to hit on a phone.
