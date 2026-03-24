---
description: Create a comprehensive strategic plan with structured task breakdown
argument-hint: Task-Name  (e.g., "refactor authentication system", "implement microservices")
---

You are an elite strategic planning specialist. Create a comprehensive, actionable plan for the task named: $ARGUMENTS

## Instructions

1. **Analyze the request** documented in `docs/tasks/[task-name]/prd.md` and determine the scope of planning needed
2. **Examine relevant files** in the codebase to understand current state
3. **Create a structured plan** with:
   - Executive Summary
   - Current State Analysis
   - Proposed Future State
   - Implementation Phases (broken into sections)
   - Detailed Tasks (actionable items with clear acceptance criteria)

   - Risk Assessment and Mitigation Strategies
   - Success Metrics
   - Required Resources and Dependencies
   - Timeline Estimates


4. **Task Breakdown Structure**:
   - Each major section represents a phase or component
   - Number and prioritize tasks within sections
   - Include clear acceptance criteria for each task
   - Specify dependencies between tasks
   - Estimate effort levels (S/M/L/XL)

5. **Create task management structure**:
   - Open directory: `docs/tasks/[task-name]/` (relative to project root)
   - Generate three files:
     - `[task-name]-plan.md` - The comprehensive plan
     - `[task-name]-context.md` - Key files, decisions, dependencies
     - `[task-name]-tasks.md` - Checklist format for tracking progress
   - Include "Last Updated: YYYY-MM-DD" in each file

## Quality Standards
- Plans must be self-contained with all necessary context
- Use clear, actionable language
- Include specific technical details where relevant
- Consider both technical and business perspectives
- Account for potential risks and edge cases

## Context References
- Check `/docs/architecture.md` for architecture overview (if exists)
- Use `docs/tasks/README.md` for task management guidelines (if exists)

**Note**: This command is ideal to use AFTER exiting plan mode when you have a clear vision of what needs to be done. It will create the persistent task structure that survives context resets.
