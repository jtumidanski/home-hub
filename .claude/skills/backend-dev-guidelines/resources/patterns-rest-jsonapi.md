
---
title: REST and JSON:API Pattern
description: Handler and transport conventions for JSON:API-compliant endpoints.
---

# REST and JSON:API Pattern

## Principles
- Handlers are thin and validate first.
- Use `server.MarshalResponse` for success.
- Map domain errors to HTTP status in a helper.

## Handler Example
```go
func createHandler(db *gorm.DB) rest.InputHandler[CreateRequest] {
  return func(d *rest.HandlerDependency, c *rest.HandlerContext, req CreateRequest) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
      processor := NewProcessor(d.Logger(), d.Context(), db)
      result, err := processor.Create(req.Data.Attributes)()
      if err != nil {
        writeErrorResponse(w, http.StatusConflict, err.Error())
        return
      }
      server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(result)
    }
  }
}
```

## Error Helper
```go
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(statusCode)
  _ = json.NewEncoder(w).Encode(map[string]any{
    "error": map[string]any{
      "status": statusCode,
      "title":  http.StatusText(statusCode),
      "detail": message,
    },
  })
}
```

## Validation Guidelines
- Required fields and header presence
- 400 for malformed input; 409 for business rule violations
