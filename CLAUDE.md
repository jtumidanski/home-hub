# Home Hub

## Project Overview

This is a Go microservices project. The primary language is Go. TypeScript is used only for  the front end ui. Always verify Docker builds when changing shared libraries.

## Workflow Rules

When asked to understand or plan something, DO NOT start implementing code changes. Wait for explicit approval before making any edits. Planning and implementation are separate phases.

## Build & Verification

After making changes across multiple services, always run builds and tests for ALL affected services before reporting completion. Expect multiple fix-and-rebuild cycles for large refactors.

## Local Deployment

Use `scripts/local-up.sh` to build and start all services locally via Docker Compose. It handles the `.env` file and build context automatically.

## Code Patterns

When refactoring shared types or creating common libraries, prefer straightforward moves over re-exporting type aliases. Keep abstractions clean — don't break service boundaries by having one layer call another's internals directly.
