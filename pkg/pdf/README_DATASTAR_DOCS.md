# Datastar v1.0 Documentation from .src/northstar

This directory now contains comprehensive documentation about Datastar v1.0 syntax extracted from the `.src/northstar` reference implementation.

## Files Included

### 1. DATASTAR_V1_SYNTAX_GUIDE.md
**Comprehensive guide** covering all Datastar v1.0 features with detailed examples:

- Initialize Signals/Store (SSE and inline)
- Bind Form Inputs (basic and named)
- Handle Form Submissions (click, keyboard, complex validation)
- GET/POST/PUT/DELETE Requests (with path parameters)
- Text Binding ($signal syntax)
- Event Handlers (click, click-outside, keydown, custom events)
- Attribute Binding (data-attr:, dynamic attributes)
- Web Components Integration
- Complete Todo App Example

**Use this when:** You need detailed explanations and full code examples for any Datastar feature.

---

### 2. DATASTAR_QUICK_REFERENCE.md
**Quick lookup card** with syntax snippets and patterns:

- Signal initialization patterns
- Form input binding syntax
- HTTP request helpers
- Event handlers
- Text & attribute binding
- Common patterns (validation, loading states, list items, filters)
- Attribute cheat sheet (table format)

**Use this when:** You know what you want to do but need the exact syntax.

---

### 3. SOURCE_FILES_REFERENCE.md
**Directory** of all template files in `.src/northstar` with:

- File locations and descriptions
- Key patterns in each file
- Code snippets from actual source
- Recommendations for which file to look at for specific features

**Use this when:** You want to find working examples in the actual source code.

---

## Quick Start

### If you're new to Datastar:
1. Start with **DATASTAR_V1_SYNTAX_GUIDE.md** - read the sections in order
2. Look at **source files** in ORDER (counter → reverse → todo → monitor)
3. Keep **DATASTAR_QUICK_REFERENCE.md** nearby for syntax lookup

### If you know Datastar basics:
1. Check **DATASTAR_QUICK_REFERENCE.md** for the syntax you need
2. Use **SOURCE_FILES_REFERENCE.md** to find a working example
3. Refer to **DATASTAR_V1_SYNTAX_GUIDE.md** for detailed explanations

### If you need a specific pattern:
Use this lookup table:

| Pattern | File | Section |
|---------|------|---------|
| Initialize signals | QUICK_REF or SYNTAX_GUIDE | "Signal Initialization" |
| Bind input | QUICK_REF or counter.templ | "Form Input Binding" |
| Click handler | QUICK_REF or counter.templ | "Event Handlers" |
| Form submission | todo.templ | "TodoInput" component |
| POST request | QUICK_REF or counter.templ | "HTTP Requests" |
| Text display | QUICK_REF or counter.templ | "Text & Attribute Binding" |
| Web component | reverse.templ or sortable.templ | "Web Components Integration" |
| Real-time updates | monitor.templ | "SSE streaming" |
| Loading state | todo.templ | "Toggle All button" |
| List rendering | todo.templ | "TodoRow component" |
| Validation | todo.templ | "TodoInput" component |

---

## Key Concepts

### Signals
Client-side reactive state that updates automatically when changed. Access via `$signalName`.

```html
<!-- Initialize -->
<div data-signals={count:0, name:''}>

<!-- Use in binding -->
<input data-bind:count />
<div data-text="$count"></div>
```

### Data Attributes
Everything uses `data-*` attributes (unlike HTMX which uses `hx-*`):

- `data-init` - Initialize with SSE data
- `data-signals` - Set initial signals
- `data-bind:field` - Two-way input binding
- `data-on:event` - Event handlers
- `data-text` - Display signals
- `data-attr:name` - Bind to attributes

### Server-Sent Events (SSE)
Bidirectional communication channel - client sends requests, server sends back partial HTML and signal updates.

```html
<div data-init={ datastar.GetSSE("/api/endpoint") }></div>
```

### HTTP Helpers
The `datastar-go` package provides type-safe Go functions:

```go
datastar.GetSSE("/api/endpoint")
datastar.PostSSE("/api/endpoint/%d", id)
datastar.PutSSE("/api/endpoint")
datastar.DeleteSSE("/api/endpoint/%d", id)
```

---

## Source Files in .src/northstar

All reference files are located at:
`/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/northstar/features/`

**Recommended reading order:**

1. **counter.templ** - Start here (signals + text binding)
2. **reverse.templ** - Next (input binding + web components)
3. **todo.templ** - Most comprehensive (complete patterns)
4. **monitor.templ** - Real-time updates (SSE streaming)
5. **sortable.templ** - Advanced (complex components)

---

## Common Patterns

### Pattern 1: Simple Signal Display
```html
<div data-signals={count:0}>
    <button data-on:click={ datastar.PostSSE("/inc") }>
        Increment
    </button>
    <div data-text="$count"></div>
</div>
```

### Pattern 2: Form Input with Submission
```html
<input
    data-bind:input
    data-on:keydown={`
        if (evt.key !== 'Enter') return;
        ${datastar.PutSSE("/submit")};
        $input = '';
    `}
/>
```

### Pattern 3: List with CRUD
```html
<ul>
    for i, item := range items {
        <li>
            <span data-text="$item.name"></span>
            <button data-on:click={ datastar.DeleteSSE("/items/%d", i) }>
                Delete
            </button>
        </li>
    }
</ul>
```

### Pattern 4: Web Component Integration
```html
<my-component
    data-signals={value:''}
    data-attr:title="$title"
    data-on:change="$value = evt.detail"
></my-component>
```

### Pattern 5: Loading State
```html
<button
    data-on:click={ datastar.PostSSE("/action") }
    data-indicator="loading"
    data-attrs-disabled="$loading"
>
    Submit
</button>
```

---

## Differences from HTMX

| Feature | HTMX | Datastar |
|---------|------|----------|
| Directives | `hx-*` | `data-*` |
| Client State | None (stateless) | Signals (reactive) |
| Input Binding | No | Yes (`data-bind:`) |
| Web Components | Limited | First-class |
| Real-time | Manual SSE setup | Built-in SSE |
| API | Multiple directives | Single unified API |

---

## Next Steps

1. **Read** DATASTAR_V1_SYNTAX_GUIDE.md cover-to-cover for complete understanding
2. **Reference** DATASTAR_QUICK_REFERENCE.md when writing code
3. **Study** the actual template files in `.src/northstar/features/`
4. **Build** your own component using these patterns

---

## Resources

- **Datastar Docs:** Check the official Datastar documentation
- **Source Code:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/northstar/`
- **Templ:** https://templ.guide/ - for template syntax
- **datastar-go:** GitHub package used for server-side helpers

---

Generated from `.src/northstar` reference implementation
Date: November 2024
