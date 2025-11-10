# Datastar v1.0 Syntax Documentation

Welcome! This directory contains comprehensive documentation about Datastar v1.0 syntax extracted from working examples in `.src/northstar`.

## Quick Navigation

### I'm new to Datastar
1. Start with **README_DATASTAR_DOCS.md** (overview & concepts)
2. Look at the source files in this order:
   - `counter.templ` - Basic signals
   - `reverse.templ` - Input binding
   - `todo.templ` - Complete example
3. Refer to **DATASTAR_QUICK_REFERENCE.md** for syntax

### I know what I want to do
1. Check **DATASTAR_QUICK_REFERENCE.md** for syntax
2. Find working code in **SOURCE_FILES_REFERENCE.md**
3. Read detailed explanation in **DATASTAR_V1_SYNTAX_GUIDE.md**

### I need the complete reference
Read **DATASTAR_V1_SYNTAX_GUIDE.md** - it has everything with examples.

## Files in This Directory

| File | Size | Purpose |
|------|------|---------|
| **START_HERE.md** | this file | Navigation guide |
| **README_DATASTAR_DOCS.md** | 6.5 KB | Overview, quick start, patterns |
| **DATASTAR_V1_SYNTAX_GUIDE.md** | 15 KB | Complete reference with examples |
| **DATASTAR_QUICK_REFERENCE.md** | 5.2 KB | Syntax cheat sheet |
| **SOURCE_FILES_REFERENCE.md** | 9.2 KB | Working examples from source |
| **DATASTAR_DOCS_INDEX.txt** | 7.2 KB | Index of all content |

## What You'll Learn

### 1. Initialize Signals/Store
```html
<!-- Fetch from server via SSE -->
<div data-init={ datastar.GetSSE("/api/endpoint") }></div>

<!-- Initialize with Go struct -->
<div data-signals={ templ.JSONString(signals) }></div>

<!-- Inline JSON -->
<div data-signals={count:0, name:''}></div>
```

### 2. Bind Form Inputs
```html
<!-- Two-way binding -->
<input data-bind:input />

<!-- Named signal binding -->
<input data-bind:_name />

<!-- Access in handlers -->
<input data-on:change={ datastar.PostSSE("/api/update") } />
```

### 3. Handle Form Submissions
```html
<!-- Click handler -->
<button data-on:click={ datastar.PostSSE("/api/action") }>
    Action
</button>

<!-- Keydown with validation -->
<input
    data-bind:input
    data-on:keydown={`
        if (evt.key !== 'Enter' || !$input.trim()) return;
        datastar.PutSSE('/api/submit');
        $input = '';
    `}
/>
```

### 4. GET/POST/PUT/DELETE Requests
```html
<!-- GET - fetch data -->
<div data-init={ datastar.GetSSE("/api/items") }></div>

<!-- POST - create/action -->
<button data-on:click={ datastar.PostSSE("/api/items") }>Add</button>

<!-- PUT - update -->
<button data-on:click={ datastar.PutSSE("/api/items/%d", id) }>Update</button>

<!-- DELETE - remove -->
<button data-on:click={ datastar.DeleteSSE("/api/items/%d", id) }>Delete</button>
```

### 5. Display Signals
```html
<!-- Text binding -->
<div data-text="$count"></div>
<span data-text="$user.name"></span>

<!-- Attribute binding -->
<component data-attr:title="$title"></component>

<!-- Dynamic disable -->
<button data-attrs-disabled="$loading">Submit</button>
```

### 6. Event Handlers
```html
<!-- Click -->
<button data-on:click={ datastar.PostSSE("/inc") }>+</button>

<!-- Keyboard -->
<input data-on:keydown={ handler } />

<!-- Click outside -->
<input data-on:click__outside={ datastar.PutSSE("/cancel") } />

<!-- Custom events -->
<my-component data-on:change="$signal = evt.detail"></my-component>
```

### 7. Web Components
```html
<my-component
    data-signals={value:'', list:[]}
    data-attr:title="$title"
    data-attr:items="JSON.stringify($items)"
    data-on:change="$value = evt.detail.value"
></my-component>
```

## Source Files

All examples come from these working templates:

```
.src/northstar/features/
├── counter/pages/counter.templ        - Signals 101
├── index/pages/index.templ            - SSE basics
├── index/components/todo.templ        - Complete example
├── reverse/pages/reverse.templ        - Web components
├── sortable/pages/sortable.templ      - Complex patterns
└── monitor/pages/monitor.templ        - Real-time updates
```

**Recommended study order:**
1. counter.templ (simplest)
2. reverse.templ (input binding + web components)
3. todo.templ (most features)
4. monitor.templ (SSE streaming)
5. sortable.templ (advanced)

## Key Concepts

### Signals
Client-side reactive state. Access as `$signalName` in bindings.
- Initialize with `data-signals={...}`
- Bind to inputs with `data-bind:`
- Display with `data-text=`
- Update from handlers

### Data Attributes
Everything uses `data-*` attributes (single API):
- `data-init` - Load initial data via SSE
- `data-signals` - Set initial state
- `data-bind:field` - Two-way input binding
- `data-on:event` - Event handlers
- `data-text` - Display signal value
- `data-attr:prop` - Bind to component properties

### HTTP Requests
Use `datastar-go` helpers:
- `datastar.GetSSE("/endpoint")` - Fetch data
- `datastar.PostSSE("/endpoint")` - Create/action
- `datastar.PutSSE("/endpoint")` - Update
- `datastar.DeleteSSE("/endpoint")` - Delete

All support format parameters:
```go
datastar.PostSSE("/items/%d/toggle", itemId)
```

### Server-Sent Events (SSE)
Bidirectional channel for:
1. Client sends HTTP request
2. Server sends back HTML + signal updates
3. Client updates automatically

## Common Patterns

### Pattern: Form with Validation
```html
<input
    data-bind:input
    data-on:keydown={`
        if (evt.key !== 'Enter' || !$input.trim()) return;
        ${datastar.PutSSE("/submit")};
        $input = '';
    `}
/>
```

### Pattern: Button with Loading
```html
<button
    data-on:click={ datastar.PostSSE("/action") }
    data-indicator="loading"
    data-attrs-disabled="$loading"
>
    Submit
</button>
```

### Pattern: List with Actions
```html
<ul>
    for i, item := range items {
        <li>
            <span data-text="$items[%d].name"></span>
            <button data-on:click={ datastar.PostSSE("/items/%d/toggle", i) }>
                Toggle
            </button>
            <button data-on:click={ datastar.DeleteSSE("/items/%d", i) }>
                Delete
            </button>
        </li>
    }
</ul>
```

## Next Steps

1. **If learning:** Read README_DATASTAR_DOCS.md then study source files
2. **If building:** Use DATASTAR_QUICK_REFERENCE.md + SOURCE_FILES_REFERENCE.md
3. **If confused:** Read the relevant section in DATASTAR_V1_SYNTAX_GUIDE.md

## Questions?

- **"How do I...?"** Check DATASTAR_QUICK_REFERENCE.md
- **"Show me an example"** Check SOURCE_FILES_REFERENCE.md
- **"Explain this"** Check DATASTAR_V1_SYNTAX_GUIDE.md

---

All documentation extracted from working code in `.src/northstar`

Generated November 2024
