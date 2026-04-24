# Apple Shortcuts → GitHub Actions

## Active Calories (prototype)

- **Workflow**: `.github/workflows/shortcuts-log-active-calories.yml`
- **Data file**: `data/active_calories.jsonl`

### Inputs (workflow_dispatch)

```json
{
  "value": "523.4",
  "date": "2026-04-22"
}
```

- `value` is required (float as string).
- `date` is optional; if omitted, today's date is used.

