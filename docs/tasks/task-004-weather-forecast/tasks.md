# Weather Forecast — Task Checklist

Last Updated: 2026-03-25

---

## Phase 1: Account Service — Household Location Extension

- [ ] **1.1** Extend household model with latitude, longitude, locationName fields and getters `(S)`
- [ ] **1.2** Extend household entity with nullable GORM columns and update Make() `(S)`
- [ ] **1.3** Extend household builder with setters and coordinate validation `(S)`
- [ ] **1.4** Extend household resource (RestModel, UpdateRequest, Transform) `(S)`
- [ ] **1.5** Extend household REST handlers and administrator for new fields `(S)`
- [ ] **1.6** Unit tests for location field validation, mapping, and serialization `(M)`
- [ ] **1.7** Build and verify account-service compiles and passes tests `(S)`

## Phase 2: Weather Service — Core Backend

- [ ] **2.1** Service scaffolding: go.mod, cmd/main.go, config, directory structure `(M)`
- [ ] **2.2** Open-Meteo client: forecast fetch + geocoding search + rate limiting `(M)`
- [ ] **2.3** WMO weather code mapping (code → summary + icon key) `(S)`
- [ ] **2.4** Forecast cache domain: model, entity (JSONB), builder `(M)`
- [ ] **2.5** Forecast cache: provider, administrator (upsert/delete), processor `(M)`
- [ ] **2.6** Forecast REST endpoints: /weather/current, /weather/forecast `(M)`
- [ ] **2.7** Geocoding REST endpoint: /weather/geocoding?q= `(M)`
- [ ] **2.8** Background refresh ticker goroutine `(M)`
- [ ] **2.9** Wire up main.go: config, DB, client, auth, routes, ticker, server `(S)`
- [ ] **2.10** Unit tests: weather codes, client mocks, cache logic, handlers `(L)`
- [ ] **2.11** Build and verify weather-service compiles and passes tests `(S)`

## Phase 3: Infrastructure

- [ ] **3.1** Add weather-service to go.work and run go work sync `(S)`
- [ ] **3.2** Create Dockerfile for weather-service `(S)`
- [ ] **3.3** Add weather-service to docker-compose.yml `(S)`
- [ ] **3.4** Add /api/v1/weather route to nginx.conf `(S)`
- [ ] **3.5** Create build-weather.sh and update build-all.sh `(S)`
- [ ] **3.6** Add weather-service to CI workflows (pr.yml + main.yml) `(M)`
- [ ] **3.7** Create k8s manifest (deployment + service + ingress update) `(S)`
- [ ] **3.8** Update architecture docs with weather-service `(S)`

## Phase 4: Frontend — Weather Features

- [ ] **4.1** Weather API service class (getCurrent, getForecast, searchPlaces) `(S)`
- [ ] **4.2** React Query hooks (useCurrentWeather, useWeatherForecast, useGeocodingSearch) `(S)`
- [ ] **4.3** Weather icon component (icon key → Lucide component) `(S)`
- [ ] **4.4** Dashboard weather widget (all states: loading, no-location, data, stale) `(M)`
- [ ] **4.5** Weather page with 7-day forecast `(M)`
- [ ] **4.6** Geocoding autocomplete component `(M)`
- [ ] **4.7** Integrate location search into household settings form `(S)`
- [ ] **4.8** Add /weather route to App.tsx and sidebar navigation `(S)`
- [ ] **4.9** Frontend component tests `(M)`
- [ ] **4.10** Frontend build and verify (compile + tests) `(S)`

## Final Verification

- [ ] **5.1** Full local stack test: docker-compose up, set location, verify weather displays `(M)`
- [ ] **5.2** Verify cache refresh cycle works (wait for ticker, confirm updated fetchedAt) `(S)`
- [ ] **5.3** Verify cache invalidation on location change `(S)`
- [ ] **5.4** All services build: `./scripts/build-all.sh` `(S)`
- [ ] **5.5** All tests pass: `./scripts/test-all.sh` `(S)`
- [ ] **5.6** Create weather-service documentation (domain.md, rest.md, storage.md) `(M)`
