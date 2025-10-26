# wellknown

**Universal Go library for generating and opening deep links across the Google and Apple app ecosystems.**  
Pure Go · Zero deps · Deterministic URLs · Cross-platform.

---

## ✨ Overview

`wellknown` lets Go applications and CLIs create **native deep links** and **URL schemes** for common apps such as:

| Category | Google | Apple |
|-----------|---------|--------|
| Calendar | `googlecalendar://render?...` | `calshow:` |
| Maps | `comgooglemaps://?q=` | `maps://?q=` |
| Mail | `mailto:` | `mailto:` |
| Drive / Files | `googledrive://` | `shareddocuments://` |

The library also provides safe fallbacks to open the **web equivalents** when native apps aren’t available.

---

## 🧩 Features

- ✅ **Pure Go** — no external dependencies.  
- 🧠 **Deterministic**: same input → same output (great for reproducible infra / NATS messages).  
- ⚙️ **Cross-platform**: works on macOS, Windows, Linux, iOS, and Android.  
- 🕹 **Programmatic & CLI**: embed in binaries or call from shell scripts.  
- 🔗 **App-aware**: automatically chooses local URL scheme vs. browser fallback.  

---

## 🧱 Installation

```bash
go get github.com/joeblew999/wellknown
```

## Examples

There are 2 examples in the examples folder.


