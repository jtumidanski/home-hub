# Task 056 — pre-PR review fixes

Source: `audit.md` § Backend guidelines (verdict NEEDS-WORK), `audit-backend.json`.
Branch state at hand-off: HEAD `0d09f97` + this docs commit; flagless `tools/verify.sh`
passed at `0d09f97`. Plan adherence: no code gaps. Frontend guidelines: PASS.

User decision: fix DOM-20 before opening the PR.

## Fix 1 — DOM-20: convert new tests to table-driven form (blocking)

Convert the new tests to table-driven `[]struct{...}` + `t.Run(tc.name, ...)` form where the
cases share a setup and assertion shape. Do not force tests with genuinely distinct
setup/assertions into one table, and do not change production code or test coverage.

### Files

- `services/recipe-service/internal/recipeimage/resource_test.go` — primary candidates
  `TestPutImageErrors`, `TestGetImageErrors`, `TestParseSize` (audit cites lines 83, 118,
  164, 189, 212, 228, 242)
- `services/recipe-service/internal/recipeimage/processor_test.go` — lines 62-255, incl.
  `TestOpenMapsErrors`
- `services/recipe-service/internal/objectstore/` `*_test.go`
- `services/recipe-service/internal/recipe/visibility_test.go`
- `services/recipe-service/internal/recipe/image_enrichment_test.go`
- the retention tests touched by this branch
  (`git diff --name-only c55e085..HEAD -- 'services/recipe-service/**/*_test.go'`)

Pattern to copy: `services/recipe-service/internal/recipeimage/imaging/imaging_test.go:43`
and `imaging/orientation_test.go:19` (existing tables on this branch).

Keep the review-focus pins intact and named: `TestImageEnrichmentOtherHousehold`,
`TestProcessOrientedDownscale`, `TestProcessPalettedPNGWithTransparency`,
`TestGetImageNotModifiedWeakList` (a pin may become a named case `t.Run` subtest only if
its top-level name still appears in `go test -run` output).

Verify: `cd services/recipe-service && go build ./... && go test ./... -count=1`, then the
flagless `tools/verify.sh` (task-verifier) must exit 0.

## Fix 2 — DOM-05: no `TransformSlice` in `recipeimage/rest.go` (waived)

`recipe-images` is a singleton per recipe with no list endpoint; a `TransformSlice` would be
dead code. Waived — no change.

## Non-blocking (not required before PR)

- Backend DOM-08/09/11/13 warnings — see audit.md.
- Frontend: bare `catch {}` at `RecipeFormPage.tsx:144`; disabled drop zone stays focusable
  (`recipe-image-field.tsx` ~91-96); dead `if (!savedId) return;` (~149).
- Final review: reaper batch ignores `attempts` (poison rows can starve newer ones);
  `design.md` still says MinIO; `.env.example` `S3_ENDPOINT` comment.

## Still open before merge (user-owned)

Manual compose acceptance items 2-9 and the UI visual check in `verification.md`.

## After the fix

Re-run the guideline review for the backend only (`backend-guidelines-reviewer`, scoped to
the fix range), then push `task-056-recipe-images` and open the PR against `main`. Mention
in the PR body: the `.gitignore` `recipes/` → `/recipes/` anchor, the RustFS swap, and the
open acceptance items.
