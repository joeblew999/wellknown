# Datastar v1.0 Source Files Reference

This document lists all the template files in `.src/northstar` where Datastar v1.0 syntax examples can be found.

## Directory Structure

```
.src/northstar/
├── features/
│   ├── index/
│   │   ├── pages/index.templ
│   │   └── components/todo.templ
│   ├── counter/
│   │   ├── pages/counter.templ
│   ├── reverse/
│   │   ├── pages/reverse.templ
│   ├── sortable/
│   │   ├── pages/sortable.templ
│   ├── monitor/
│   │   ├── pages/monitor.templ
│   └── common/
│       ├── components/shared.templ
│       └── layouts/base.templ
```

## File Summaries

### index/pages/index.templ
**Location:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/northstar/features/index/pages/index.templ`

**Key Pattern:** SSE Initialization
```html
<div id="todos-container" data-init={ datastar.GetSSE("/api/todos") }></div>
```

**Shows:**
- Initial data loading via SSE
- Container with SSE data handler

---

### counter/pages/counter.templ
**Location:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/northstar/features/counter/pages/counter.templ`

**Key Patterns:**
1. Signal initialization from Go struct
2. Text binding to signals
3. Click event handlers with POST requests

**Code Examples:**
```go
type CounterSignals struct {
    Global uint32 `json:"global"`
    User   uint32 `json:"user"`
}

templ Counter(signals CounterSignals) {
    <div
        id="container"
        data-signals={ templ.JSONString(signals) }
        class="flex flex-col gap-4"
    >
        @CounterButtons()
        @CounterCounts()
    </div>
}
```

**Shows:**
- Struct-based signal initialization
- Data binding in Templ
- Multiple text bindings
- Click handlers with POST

---

### reverse/pages/reverse.templ
**Location:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/northstar/features/reverse/pages/reverse.templ`

**Key Patterns:**
1. Input binding with data-bind:
2. Web component event handling
3. Attribute binding to web components
4. Signal synchronization with custom events

**Code Examples:**
```html
<input type="text" data-bind:_name=""/>

<p class="truncate" data-signals:_reversed="" data-text="$_reversed"></p>

<reverse-component 
    data-on:reverse="$_reversed = evt.detail.value"
    data-attr:name="$_name"
></reverse-component>
```

**Shows:**
- Two-way input binding
- Web component integration
- Custom event handling
- Signal to attribute binding
- Custom element communication

---

### index/components/todo.templ
**Location:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/northstar/features/index/components/todo.templ`

**Key Patterns:**
1. Complex form handling
2. Conditional rendering
3. Dynamic signal binding
4. Multiple event handlers
5. Indicator/loading states
6. Attribute disabling

**Code Examples:**

Input binding with keydown handler:
```html
<input
    id="todoInput"
    data-bind:input
    data-on:keydown={ fmt.Sprintf(`
        if (evt.key !== 'Enter' || !$input.trim().length) return;
        %s;
        $input = '';
    `, datastar.PutSSE("/api/todos/%d/edit", i)) }
/>
```

Dynamic signals:
```html
<div
    class="flex flex-col w-full gap-4"
    data-signals={ fmt.Sprintf("{input:'%s'}", input) }
>
```

Button with indicator:
```html
<button
    id="toggleAll"
    class="btn btn-lg"
    data-on:click={ datastar.PostSSE("/api/todos/-1/toggle") }
    data-indicator="toggleAllFetching"
    data-attrs-disabled="$toggleAllFetching"
>
    Toggle All
</button>
```

List rendering with event handlers:
```html
<ul class="divide-y divide-primary">
    for i, todo := range mvc.Todos {
        <li class="flex items-center gap-8 p-2" id={ fmt.Sprintf("todo%d", i) }>
            <label
                data-on:click={ datastar.PostSSE("/api/todos/%d/toggle", i) }
                data-indicator={ fetchingSignalName }
            >
                Checkbox Icon
            </label>
            
            <label
                data-on:click={ datastar.GetSSE("/api/todos/%d/edit", i) }
            >
                { todo.Text }
            </label>
            
            <button
                data-on:click={ datastar.DeleteSSE("/api/todos/%d", i) }
                data-indicator={ fetchingSignalName }
                data-attrs-disabled={ fetchingSignalName + "" }
            >
                Delete
            </button>
        </li>
    }
</ul>
```

Filter buttons:
```html
<div class="join">
    for i := TodoViewModeAll; i < TodoViewModeLast; i++ {
        if i == mvc.Mode {
            <div class="btn btn-xs btn-primary join-item">
                { TodoViewModeStrings[i] }
            </div>
        } else {
            <button
                class="btn btn-xs join-item"
                data-on:click={ datastar.PutSSE("/api/todos/mode/%d", i) }
            >
                { TodoViewModeStrings[i] }
            </button>
        }
    }
</div>
```

**Shows:**
- Complete form submission patterns
- Validation in handlers
- Dynamic signal creation
- Multiple request types (GET/POST/PUT/DELETE)
- Loading indicators
- Dynamic attribute binding
- Conditional event handlers
- List rendering with events

---

### sortable/pages/sortable.templ
**Location:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/northstar/features/sortable/pages/sortable.templ`

**Key Patterns:**
1. Complex signal initialization
2. Multiple attribute bindings
3. JSON serialization in attributes
4. Custom event handling

**Code Examples:**
```html
<sortable-example
    data-signals="{title: 'Item Info', info:'', items: [{name: `item one`}, ...]}"
    data-attr:title="$title"
    data-attr:value="$info"
    data-attr:items="JSON.stringify($items)"
    data-on:change="event.detail && console.log(`...`)"
></sortable-example>
```

**Shows:**
- Inline signal initialization with nested objects
- Multiple attribute bindings
- JSON serialization
- Custom event logging
- Web component with complex data

---

### monitor/pages/monitor.templ
**Location:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/northstar/features/monitor/pages/monitor.templ`

**Key Patterns:**
1. Real-time SSE data streaming
2. Multiple signals from server
3. Text binding for real-time updates

**Code Examples:**
```go
type SystemMonitorSignals struct {
    MemTotal       string `json:"memTotal,omitempty"`
    MemUsed        string `json:"memUsed,omitempty"`
    MemUsedPercent string `json:"memUsedPercent,omitempty"`
    CpuUser        string `json:"cpuUser,omitempty"`
    CpuSystem      string `json:"cpuSystem,omitempty"`
    CpuIdle        string `json:"cpuIdle,omitempty"`
}

templ MonitorPage() {
    <div
        id="container"
        data-init={ datastar.GetSSE("/monitor/events") }
        data-signals="{memTotal:'', memUsed:'', memUsedPercent:'', cpuUser:'', cpuSystem:'', cpuIdle:''}"
    >
        <div id="mem" class="flex flex-col">
            <p>Total: <span data-text="$memTotal"></span></p>
            <p>Used: <span data-text="$memUsed"></span></p>
            <p>Used (%): <span data-text="$memUsedPercent"></span></p>
        </div>
        <div id="cpu" class="flex flex-col">
            <p>User: <span data-text="$cpuUser"></span></p>
            <p>System: <span data-text="$cpuSystem"></span></p>
            <p>Idle: <span data-text="$cpuIdle"></span></p>
        </div>
    </div>
}
```

**Shows:**
- Real-time SSE streaming
- Pre-initialized signals
- Text binding for streaming updates
- Struct-based signal definitions

---

### common/layouts/base.templ
**Location:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/northstar/features/common/layouts/base.templ`

**Shows:**
- Base layout structure
- Script loading patterns
- HTML document setup

---

### common/components/shared.templ
**Location:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/northstar/features/common/components/shared.templ`

**Shows:**
- Reusable component patterns
- Navigation structure
- Common UI elements

---

## How to Use These Files

1. **For Signal Initialization:** Look at `counter.templ` and `monitor.templ`
2. **For Input Binding:** Look at `reverse.templ` and `todo.templ`
3. **For Form Submission:** Look at `todo.templ` (most comprehensive)
4. **For HTTP Requests:** Look at `todo.templ` (all 4 types shown)
5. **For Web Components:** Look at `reverse.templ` and `sortable.templ`
6. **For Real-time Updates:** Look at `monitor.templ`
7. **For Complete Example:** Look at `todo.templ` (full MVC pattern)

## Import Pattern

All templates import from the `datastar-go` package:
```go
import (
    "github.com/starfederation/datastar-go/datastar"
    ...
)
```

This provides the `datastar.GetSSE()`, `datastar.PostSSE()`, `datastar.PutSSE()`, `datastar.DeleteSSE()` helpers.

## Templ Usage

All files use the Templ templating language:
- `.templ` files are Go template components
- They export via `templ ComponentName() { ... }`
- Invoked as `@ComponentName()` or `@ComponentName(args)`
- Type-safe with Go struct support

## Environment

- **Project:** wellknown
- **Package:** github.com/joeblew999/wellknown
- **Reference Root:** `/Users/apple/workspace/go/src/github.com/joeblew999/wellknown/.src/northstar/`
