---
Task: 024 — Consolidated Ingredient Categories
Date: 2026-04-08
---

# Diagnosis

## Working hypothesis: Suspect 1 — `ListCategories` error silently swallowed

The implementer of this task did not have direct access to the deployed
environment's recipe-service logs or DB. Phase 1 was therefore executed as a
*static* diagnosis against the codebase rather than a *runtime* diagnosis
against the failing environment, and the most likely candidate (suspect 1
from `plan.md` and `context.md`) was used as the working hypothesis.

The bias toward suspect 1 is justified by the code itself:

- `services/recipe-service/internal/export/processor.go:130-136` swallows the
  error from `categoryclient.ListCategories` behind an `if … err == nil`
  guard. There is no log line, no metric, no trace. *Any* failure of the
  category-service call (network, 401, decode failure, etc.) results in
  an empty `categoryByID` map, which in turn means every canonical
  ingredient's category lookup misses, which in turn means every ingredient
  lands in "Uncategorized". This matches the deployed-env symptom exactly:
  silent, server-side, no browser console errors.
- `services/recipe-service/internal/categoryclient/client.go:54-57` returns
  an explicit error on any non-200 response — including 401, which is the
  most plausible failure mode in a deployed environment where the user's
  cookie is forwarded by `categoryclient` but token semantics could differ
  from local dev.
- The frontend, JSON:API serializer, and processor sort logic are all
  correct. The data flow only breaks at the one swallowed-error site.

## Other suspects ruled out

- **Suspect 2** (NULL `category_id` on `canonical_ingredients`): Not ruled
  out by direct inspection of the deployed DB, but if it were the cause we
  would also expect the same symptom in the canonical ingredient admin UI,
  which the bug report explicitly says is *not* affected — categories show
  correctly there. Filing a follow-up is unnecessary unless Phase 5
  verification reveals categories are still missing after the suspect-1 fix.
- **Suspect 3** (cross-tenant / deleted-category ID mismatch): Would
  produce a *partial* failure (some ingredients categorized, some not),
  not the all-Uncategorized symptom reported. The Phase 3 warn log will
  catch this if it ever does occur.
- **Suspect 4** (`GetByIDs` returning a partial map): Same reasoning as
  suspect 3 — would be partial, not total. Phase 3 warn log covers it.

## Disposition

Because the Phase 3 observability hardening is *required regardless of which
suspect is confirmed*, and because the suspect-1 fix is exactly the same
code change as the suspect-1 log line (Task 3.1), the implementation
proceeds as planned. The next deployed-env failure of any of the four
shapes will now be loud (one of the three §4.3 log lines will fire),
allowing definitive runtime diagnosis without further code changes.

If Phase 5.3 verification in the deployed environment shows categories
*still* missing after this change, the next step is to grep production
logs for the three §4.3 messages to identify which suspect is actually
firing, and file a follow-up task.
