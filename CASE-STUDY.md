# Case Study: Gmail iOS Deep Linking - Deliberate Friction

> Real-world documentation of Big Tech platform lock-in discovered while trying to open a Gmail thread from an external link on iOS.

## The Problem

**Goal:** Generate a URL that opens a specific Gmail thread in the Gmail iOS app.

**Result:** Impossible. Google has not implemented this functionality.

## What We Tried

### 1. Gmail Web URL

```
https://mail.google.com/mail/u/0/#inbox/190838d0f33b71ec
```

**Result:** Opens Safari, not Gmail app. iOS doesn’t redirect because Google hasn’t registered `mail.google.com` as a Universal Link for the Gmail app.

### 2. Long-press “Open in App”

**Result:** Option doesn’t exist. Gmail app doesn’t register as a handler for Gmail web URLs.

### 3. googlegmail:// URL Scheme

```
googlegmail://co?to=someone@example.com&subject=Hello
```

**Result:** Only supports compose. No scheme exists for opening a specific thread or message.

### 4. Gmail Mobile Web → “Show Original” (to get Message-ID)

**Result:** Feature doesn’t exist in Gmail iOS app or mobile web.

### 5. Gmail Mobile Web → Request Desktop Site → “Show Original”

**Result:** Opens in new tab which reverts to mobile view. Blocked again.

## What Google *Could* Have Built

```
// Native thread deep link (doesn't exist)
googlegmail://thread/190838d0f33b71ec

// Native message deep link (doesn't exist)
googlegmail://message/19ae4d662e46ed15

// Universal Link registration (not implemented)
// mail.google.com URLs → Gmail app
```

These are trivial to implement. Google chose not to.

## Why This Is Deliberate

### Lock-in Economics

If Gmail supported proper deep linking:

- Any app could integrate with Gmail threads (CRMs, note apps, calendars)
- Gmail becomes a replaceable component in user workflows
- Users stay in *their* apps, visit Gmail only when needed
- Google loses control of the user journey

### Current State Benefits Google

- Users must **start** in Gmail to access threads
- No third-party app can provide a “click to view thread” experience
- Gmail remains a destination, not a component
- More time in Gmail = more exposure to Google’s UI, ads, upsells

### iOS Is Enemy Territory

Google and Apple are competitors. Features that make Gmail work beautifully on iOS make iPhones more attractive. On Android, Google controls the intents system. On iOS, they’re a guest with limited privileges - and they’ve chosen not to use even those.

## The iOS Mail Workaround (Potential)

iOS Mail.app has a `message://` URL scheme that can open emails by RFC822 Message-ID:

```
message://<CAJ7b2w...@mail.gmail.com>
```

**Requirements:**

1. Gmail accounts added to iOS Mail (Settings → Mail → Accounts)
1. Extract Message-ID from email headers
1. URL-encode the Message-ID
1. Generate `message://` link

**The Catch:**
Getting the Message-ID requires “Show Original” which Google blocks on iOS entirely.

**Potential Solutions:**

- Server-side: Use Gmail API with `format=raw` to extract Message-ID
- Store Message-IDs when emails are fetched
- Build a wellknown gateway that handles this translation

## Wellknown Architecture Solution

```
┌─────────────────────────────────────────────────────────────┐
│                    User clicks link                         │
│              wellknown://email/thread/abc123                │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                  Wellknown Gateway                          │
│                                                             │
│  1. Lookup thread abc123 in local Automerge store           │
│  2. Get Message-ID from stored metadata                     │
│  3. Detect platform                                         │
└─────────────────────┬───────────────────────────────────────┘
                      │
          ┌───────────┼───────────┐
          │           │           │
          ▼           ▼           ▼
     ┌────────┐  ┌────────┐  ┌────────┐
     │  iOS   │  │Android │  │Desktop │
     │        │  │        │  │        │
     │message:│  │intent: │  │https://│
     │//msgid │  │gmail   │  │mail... │
     └────────┘  └────────┘  └────────┘
```

### Key Insight

By owning your email data (via IMAP sync to Automerge), you have access to Message-IDs that Google deliberately hides on mobile. Your gateway can generate working deep links that Google’s own URLs cannot provide.

**This inverts the power dynamic:**

- Google’s URLs → broken on iOS
- Wellknown URLs → work everywhere (because you own the metadata)

## Implementation TODO

- [ ] Add `email` package to wellknown
- [ ] Implement Message-ID extraction from Gmail API (format=raw)
- [ ] Store Message-IDs in Automerge alongside email metadata
- [ ] Add iOS Mail `message://` scheme generation
- [ ] Add Android Gmail intent generation
- [ ] Test multi-account scenarios (3 Gmail accounts)
- [ ] Document iOS Mail account setup requirements

## Code Sketch

```go
package email

import (
    "fmt"
    "net/url"
    "runtime"
)

type Platform string

const (
    PlatformIOS     Platform = "ios"
    PlatformAndroid Platform = "android"
    PlatformDesktop Platform = "desktop"
)

type ThreadLink struct {
    ThreadID  string
    MessageID string // RFC822 Message-ID, e.g., <CAJ7b2w...@mail.gmail.com>
    WebURL    string
}

// GenerateDeepLink returns a platform-appropriate URL for opening an email thread
func (t *ThreadLink) GenerateDeepLink(platform Platform) string {
    switch platform {
    case PlatformIOS:
        if t.MessageID != "" {
            // iOS Mail can open by Message-ID
            encoded := url.PathEscape(t.MessageID)
            return fmt.Sprintf("message://%s", encoded)
        }
        // Fallback to web (will open Safari, not Gmail app)
        return t.WebURL
        
    case PlatformAndroid:
        // Android intent might work better
        // TODO: Test gmail intent with thread ID
        return fmt.Sprintf("intent://mail.google.com/mail/#inbox/%s#Intent;scheme=https;package=com.google.android.gm;end", t.ThreadID)
        
    case PlatformDesktop:
        return t.WebURL
        
    default:
        return t.WebURL
    }
}
```

## Related Reading

- [Apple URL Schemes Reference](https://developer.apple.com/documentation/xcode/defining-a-custom-url-scheme-for-your-app)
- [Android Intents and Intent Filters](https://developer.android.com/guide/components/intents-filters)
- [RFC 2392 - Content-ID and Message-ID URLs](https://datatracker.ietf.org/doc/html/rfc2392)

## Conclusion

This case study demonstrates why wellknown exists. Big Tech platforms deliberately cripple interoperability to maintain lock-in. The solution isn’t to fight their URL schemes - it’s to own your data and generate your own links that work despite their friction.

When you own the metadata, you own the links.

-----

*Documented: December 2024*
*Context: Attempting to share a Gmail thread link that opens correctly on iOS*
