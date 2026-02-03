---
name: scribe-development
description: Scribe app development guide for Wails v3 + Vue 3 desktop app. Use when building, debugging, or modifying the Scribe plugin manager.
---

# Scribe Development

## Project Structure

- Frontend: Vue 3 + TypeScript in `frontend/`
- Backend: Go with Wails v3 bindings in `main.go` and `internal/`
- Packaging: Platform-specific scripts in `packaging/`

## Building

Build CLI binary:
```bash
make build
# Output: build/bin/scribe
```

Build macOS app bundle:
```bash
cp build/bin/scribe build/scribe && ./packaging/macos/create-app.sh
# Output: build/Scribe.app
```

Note: `create-app.sh` expects binary at `build/scribe`, not `build/bin/scribe`.

## Wails v3 Dialogs

Browser APIs like `confirm()`, `alert()`, `prompt()` are silently blocked in Wails apps.

### Option 1: Inline Confirmation (Preferred for UX)

Use Vue reactive state to show/hide an inline confirmation UI:

```typescript
const pendingAction = ref<string | null>(null)

function handleClick(item: string) {
  pendingAction.value = item  // Show confirmation
}

function cancel() {
  pendingAction.value = null
}

async function confirm() {
  const item = pendingAction.value
  pendingAction.value = null
  await performAction(item)
}
```

```vue
<div v-if="pendingAction" class="confirm-dialog">
  <p>Confirm action on "{{ pendingAction }}"?</p>
  <button @click="cancel">Cancel</button>
  <button @click="confirm">Confirm</button>
</div>
```

### Option 2: Wails Runtime Dialogs

For native OS dialogs:

```typescript
import { Dialogs } from '@wailsio/runtime'

const result = await Dialogs.Question({
  Title: 'Confirm',
  Message: 'Are you sure?',
  Buttons: [
    { Label: 'Yes', IsDefault: true },
    { Label: 'No', IsCancel: true }
  ]
})
if (result === 'Yes') { ... }
```

## wails3 CLI

If `wails3` not found:
```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

Do NOT use `wails3 generate build-assets` - it creates unnecessary files for all platforms. This project uses Makefile + packaging scripts instead.

## Debugging

Add console logs in frontend components:
```typescript
console.log('[ComponentName] action', data)
```

View in browser dev tools: Cmd+Option+I in the Wails app window.

## Key Files

| Purpose | Location |
|---------|----------|
| Frontend source | `frontend/src/` |
| Wails bindings | `frontend/src/bindings/` |
| Go backend | `main.go`, `internal/` |
| macOS packaging | `packaging/macos/create-app.sh` |
