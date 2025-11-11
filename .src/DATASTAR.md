# Datastar Implementation Guide

Choose your approach based on project needs:

## 1. Via Framework (Recommended for new projects)
**Path**: `.src/via/internal/examples/`
- Pure Go, no templates needed
- Uses `via/h` package for HTML composition (gomponents)
- Built-in signals, actions, SSE handling
- Example: `.src/via/internal/examples/counter/main.go`

```go
v.Page("/", func(c *via.Context) {
    data := Counter{Count: 0}
    step := c.Signal(1)
    increment := c.Action(func() {
        data.Count += step.Int()
        c.Sync()
    })
    c.View(func() h.H {
        return h.Div(
            h.P(h.Textf("Count: %d", data.Count)),
            h.Button(h.Text("Increment"), increment.OnClick()),
        )
    })
})
```

## 2. Northstar Pattern (Go + Templ)
**Path**: `.src/northstar/features/counter/`
- Go handlers + templ templates
- Manual datastar integration
- More control, more setup

```templ
<button data-on:click={datastar.PostSSE("/counter/increment")}>
    Increment
</button>
```

## 3. DatastarUI Components (Component library)
**Path**: `.src/datastarui/components/`
- Pre-built shadcn-style components
- Templ-based
- Best for using existing components, not building from scratch

## 4. Low-Level SDK (Advanced)
**Path**: `.src/datastar-go/datastar/`
- Direct SDK usage
- Manual SSE, signals, elements
- Use when you need full control

## Quick Start Recommendation

For simple reactive apps: **Use Via**
- Minimal setup
- No template compilation
- Fast iteration

For complex UIs with components: **Use DatastarUI + Northstar patterns**
- Reusable components
- Template-based
- More boilerplate
