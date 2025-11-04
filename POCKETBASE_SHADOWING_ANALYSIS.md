# PocketBase Field Shadowing Analysis

## Executive Summary

The PocketBase codebase has **critical shadowing issues** where generated record proxy field getters/setters conflict with built-in `core.Record` methods. The **email**, **password**, and **tokenKey** fields are the primary culprits, but many other methods on `core.Record` could also be shadowed.

**Key Finding**: Renaming fields (e.g., `email` → `userEmail`) is NOT the recommended solution. Instead, PocketBase's code generation explicitly prevents this by detecting shadowing at generation time and failing the build with a clear error message.

---

## 1. How core.Record and BaseRecordProxy Work

### 1.1 BaseRecordProxy Structure

```go
// From: .src/pocketbase/core/record_proxy.go

type RecordProxy interface {
    ProxyRecord() *Record
    SetProxyRecord(record *Record)
}

type BaseRecordProxy struct {
    *Record  // embedded pointer to the proxied Record
}

func (m *BaseRecordProxy) ProxyRecord() *Record {
    return m.Record
}

func (m *BaseRecordProxy) SetProxyRecord(record *Record) {
    m.Record = record
}
```

**How it works:**
- Generated proxy structs embed `BaseRecordProxy`
- `BaseRecordProxy` embeds `*Record` (pointer embedding)
- This means all methods and fields from `*Record` are automatically promoted to the proxy
- Generated code accesses fields via methods like `p.Email()` which get proxified to `p.Get("email")`

### 1.2 Public Record Fields

The `core.Record` struct has only ONE public field:
- `Id` - The primary key

All other data is stored in private fields:
- `data` - the current field values (store.Store)
- `originalData` - original database values
- `expand` - relationship expansions
- `collection` - reference to the collection schema

---

## 2. Critical Methods on core.Record That Shadow Generated Code

### 2.1 Auth-Related Methods (Most Critical)

Located in `.src/pocketbase/core/record_model_auth.go`:

```go
// Email() method - CRITICAL SHADOWING
func (m *Record) Email() string {
    return m.GetString(FieldNameEmail)
}

// SetEmail() method
func (m *Record) SetEmail(email string) {
    m.Set(FieldNameEmail, email)
}

// TokenKey() method - CRITICAL SHADOWING
func (m *Record) TokenKey() string {
    return m.GetString(FieldNameTokenKey)
}

// SetTokenKey() method
func (m *Record) SetTokenKey(key string) {
    m.Set(FieldNameTokenKey, key)
}

// Password handling (SetPassword, ValidatePassword, SetRandomPassword)
func (m *Record) SetPassword(password string) {
    m.Set(FieldNamePassword, password)
}

// EmailVisibility() method
func (m *Record) EmailVisibility() bool {
    return m.GetBool(FieldNameEmailVisibility)
}

// SetEmailVisibility() method
func (m *Record) SetEmailVisibility(visible bool) {
    m.Set(FieldNameEmailVisibility, visible)
}

// Verified() method
func (m *Record) Verified() bool {
    return m.GetBool(FieldNameVerified)
}

// SetVerified() method
func (m *Record) SetVerified(verified bool) {
    m.Set(FieldNameVerified, verified)
}

// TokenKey refresh
func (m *Record) RefreshTokenKey() {
    m.Set(FieldNameTokenKey+autogenerateModifier, "")
}

// Validate password
func (m *Record) ValidatePassword(password string) bool {
    pv, ok := m.GetRaw(FieldNamePassword).(*PasswordFieldValue)
    if !ok {
        return false
    }
    return pv.Validate(password)
}
```

### 2.2 Other Shadowing-Prone Methods

From `.src/pocketbase/core/record_model.go`:

```go
// Generic field access
func (m *Record) Get(key string) any { ... }
func (m *Record) GetRaw(key string) any { ... }
func (m *Record) Set(key string, value any) { ... }
func (m *Record) SetRaw(key string, value any) { ... }

// Type-specific getters
func (m *Record) GetBool(key string) bool { ... }
func (m *Record) GetString(key string) string { ... }
func (m *Record) GetInt(key string) int { ... }
func (m *Record) GetFloat(key string) float64 { ... }
func (m *Record) GetDateTime(key string) types.DateTime { ... }
func (m *Record) GetStringSlice(key string) []string { ... }

// Expansion methods
func (m *Record) Expand() map[string]any { ... }
func (m *Record) SetExpand(expand map[string]any) { ... }
func (m *Record) MergeExpand(expand map[string]any) { ... }
func (m *Record) ExpandedOne(relField string) *Record { ... }
func (m *Record) ExpandedAll(relField string) []*Record { ... }

// Data access
func (m *Record) FieldsData() map[string]any { ... }
func (m *Record) CustomData() map[string]any { ... }

// Visibility control
func (m *Record) Hide(fieldNames ...string) *Record { ... }
func (m *Record) Unhide(fieldNames ...string) *Record { ... }

// JSON marshaling
func (m *Record) MarshalJSON() ([]byte, error) { ... }
func (m *Record) UnmarshalJSON(data []byte) error { ... }

// Collections
func (m *Record) Collection() *Collection { ... }
func (m *Record) TableName() string { ... }

// Lifecycle
func (m *Record) Original() *Record { ... }
func (m *Record) Fresh() *Record { ... }
func (m *Record) Clone() *Record { ... }
func (m *Record) IsNew() bool { ... }
func (m *Record) PostScan() error { ... }

// Flags
func (m *Record) WithCustomData(state bool) *Record { ... }
func (m *Record) IgnoreEmailVisibility(state bool) *Record { ... }
func (m *Record) IgnoreUnchangedFields(state bool) *Record { ... }
```

---

## 3. How the Shadowing Problem Manifests

### 3.1 The Problem: Field Names Shadow Methods

When you have a PocketBase schema with fields like `email`, `password`, `tokenKey`, the code generator creates:

```go
// Generated proxy struct
type User struct {
    core.BaseRecordProxy
}

// Generated getter for "email" field
func (u *User) Email() string {
    return u.GetString("email")  // Calls Record.GetString()
}

// Generated setter for "email" field
func (u *User) SetEmail(email string) {
    u.Set("email", email)  // Calls Record.Set()
}
```

**The shadowing issue**: The generated `Email()` method has the same signature as `Record.Email()`, but it calls `GetString()` while the original calls `GetString()` with the field name. This is actually COMPATIBLE, but...

### 3.2 Why This Is Actually a Problem

The problem isn't that the methods conflict in implementation—it's that:

1. **Confusion for downstream code**: Code using the proxy might call `proxy.Email()` expecting the PocketBase framework version, but gets the generated proxy version
2. **Loss of utility methods**: The PocketBase-provided methods like `SetPassword()` with automatic tokenKey refresh are shadowed
3. **Framework internals break**: Code inside PocketBase that expects `Record.Email()` to exist gets the proxy method instead if the code works with interfaces

### 3.3 Real-World Example from Tests

From `.src/pocketbase-gogen/generator/proxy_methods_test.go`:

```go
func TestSystemFields(t *testing.T) {
    template := `func (s *StructName) Method() {
        s.password = "Mb2.r5oHf-0t"
        _ = s.tokenKey
        s.tokenKey = "key"
        _ = s.email
        s.email = "test@example.com"
        _ = s.emailVisibility
        s.emailVisibility = false
        _ = s.verified
        s.verified = true
    }
    `

    expectedGeneration := `func (s *StructName) Method() {
        s.SetPassword("Mb2.r5oHf-0t")      // calls Record.SetPassword()
        _ = s.TokenKey()                     // SHADOWS Record.TokenKey()!
        s.SetTokenKey("key")                 // calls Record.SetTokenKey()
        _ = s.Email()                        // SHADOWS Record.Email()!
        s.SetEmail("test@example.com")       // calls Record.SetEmail()
        _ = s.EmailVisibility()              // SHADOWS Record.EmailVisibility()!
        s.SetEmailVisibility(false)          // calls Record.SetEmailVisibility()
        _ = s.Verified()                     // SHADOWS Record.Verified()!
        s.SetVerified(true)                  // calls Record.SetVerified()
    }
    `
}
```

The generated code **explicitly creates methods that shadow the core methods**.

---

## 4. How PocketBase's Code Generator Detects and Prevents Shadowing

### 4.1 Shadow Detection Mechanism

Located in `.src/pocketbase-gogen/generator/pb_introspection.go`:

```go
func (p *pocketBaseInfo) shadowsRecord(proxyStruct *types.Named) (bool, []string) {
    // Extract all exported names from the proxy struct
    // (including those from BaseRecordProxy and Record)
    proxyNames := extractNamesWithEmbedded(proxyStruct, p.baseProxyType)
    
    shadowed := make([]string, 0)

    // Check if any proxy names conflict with Record's names
    for name := range proxyNames {
        if _, ok := p.allRecordNames[name]; ok {
            shadowed = append(shadowed, name)  // This is a shadow!
        }
    }

    return len(shadowed) > 0, shadowed
}

func (p *pocketBaseInfo) collectRecordNames() error {
    // Dynamically extract all exported names from core.Record
    recordObj := p.pkg.Types.Scope().Lookup("Record")
    recordNamedType := recordObj.Type().(*types.Named)
    
    // This extracts:
    // - All exported methods: Email(), SetEmail(), TokenKey(), etc.
    // - All exported fields: Id
    p.allRecordNames = extractNamesWithEmbedded(recordNamedType, nil)
    return nil
}
```

The generator extracts the method set of `*core.Record` at code generation time using Go's type information system.

### 4.2 Build-Time Failure

From `.src/pocketbase-gogen/generator/generate_proxies.go`:

```go
func checkPbShadows(sourceCode []byte) error {
    // Parse generated code
    f, err := parser.ParseFile(fset, "shadowcheck.go", sourceCode, parser.SkipObjectResolution)
    
    // Get all proxy struct types
    for _, name := range names {
        obj := scope.Lookup(name)
        proxyType, _ := obj.Type().(*types.Named)
        
        // Check for shadowing
        _, shadows := pbInfo.shadowsRecord(proxyType)
        allShadows = append(allShadows, shadows...)
    }

    if len(allShadows) > 0 {
        errMsg := fmt.Sprintf(`Can not generate proxy code because some of the generated names shadow names from PocketBase's core.Record struct. This prevents the internals of PocketBase to safely handle data.
Try renaming fields/methods in the template to escape the shadowing. Don't forget to use the '// schema-name:' comment when renaming fields.
Additionally make sure that all the system fields in your template are marked by the '// system:' comment and do not change the generated system comments.
The shadowed names are: %v`, allShadows)
        return errors.New(errMsg)
    }

    return nil
}
```

**The generator FAILS THE BUILD** if shadowing is detected!

---

## 5. Why Renaming Fields (e.g., email → userEmail) Is NOT Recommended

### 5.1 The Field Renaming Approach

When you encounter the shadowing error, you might think: "Just rename the field to avoid the conflict!"

```go
// Template with renamed field
type User struct {
    core.BaseRecordProxy
}

// User field (renamed from "email" to "userEmail")
// schema-name: email
func (u *User) UserEmail() string {
    return u.GetString("email")
}
```

The `// schema-name: email` comment tells the generator to store data under the `"email"` key in the database.

### 5.2 Why This Is Wrong (According to PocketBase)

**From the error message in the generator**:

> "Try renaming fields/methods in the template to escape the shadowing. Don't forget to use the '// schema-name:' comment when renaming fields."

While this **technically works** to avoid the compilation error, it's **NOT the intended solution** because:

1. **You're hiding PocketBase's utility methods**: Methods like `Record.Email()`, `Record.SetEmail()`, `Record.TokenKey()`, `Record.RefreshTokenKey()` are designed to provide safe, framework-aware access to auth fields.

2. **You lose special behavior**: For example, `Record.SetPassword()` automatically triggers tokenKey refresh, but if your generated code calls `Record.Set("password", ...)` directly, you bypass this logic.

3. **The error message exists for a reason**: The generator explicitly prevents shadowing to ensure framework correctness. When it detects shadowing, it fails loudly rather than silently allowing broken code.

4. **Framework internals expect these methods**: PocketBase's internal code assumes these methods exist and behave correctly.

### 5.3 What PocketBase Actually Wants You To Do

**The real solution**: Don't add fields to your proxy struct that match PocketBase's built-in methods.

For auth collections, the system fields are automatically available through the `BaseRecordProxy`:
- Use `proxy.Email()` - from `Record.Email()`
- Use `proxy.SetEmail(email)` - from `Record.SetEmail()`
- Use `proxy.TokenKey()` - from `Record.TokenKey()`
- Use `proxy.SetTokenKey(key)` - from `Record.SetTokenKey()`
- Use `proxy.SetPassword(pass)` - from `Record.SetPassword()`
- Use `proxy.ValidatePassword(plain)` - from `Record.ValidatePassword()`

Don't create proxy fields that duplicate these names.

---

## 6. Complete List of Methods That Cause Shadowing

### 6.1 Auth Methods (Most Critical)

These are defined in `record_model_auth.go`:

```
Email()
SetEmail()
EmailVisibility()
SetEmailVisibility()
TokenKey()
SetTokenKey()
RefreshTokenKey()
Verified()
SetVerified()
SetPassword()
SetRandomPassword()
ValidatePassword()
```

### 6.2 Generic Data Access Methods

These are defined in `record_model.go`:

```
Get()
GetRaw()
Set()
SetRaw()
GetBool()
GetString()
GetInt()
GetFloat()
GetDateTime()
GetStringSlice()
GetGeoPoint()
GetUnsavedFiles()
GetUploadedFiles()
UnmarshalJSONField()
```

### 6.3 Expansion Methods

```
Expand()
SetExpand()
MergeExpand()
ExpandedOne()
ExpandedAll()
```

### 6.4 Record Lifecycle Methods

```
Collection()
TableName()
Original()
Fresh()
Clone()
IsNew()
PostScan()
```

### 6.5 Data Manipulation Methods

```
FieldsData()
CustomData()
WithCustomData()
IgnoreEmailVisibility()
IgnoreUnchangedFields()
Hide()
Unhide()
Load()
```

### 6.6 JSON Methods

```
MarshalJSON()
UnmarshalJSON()
```

### 6.7 File Management

```
BaseFilesPath()
FindFileFieldByFile()
```

---

## 7. Recommended Best Practices

### 7.1 DO: Use System Fields Through Proxy

```go
type User struct {
    core.BaseRecordProxy
}

// For a custom "username" field:
func (u *User) Username() string {
    return u.GetString("username")
}

func (u *User) SetUsername(username string) {
    u.Set("username", username)
}

// For auth system fields, use the embedded Record methods directly:
// u.Email()
// u.SetEmail()
// u.TokenKey()
// etc.
```

### 7.2 DON'T: Create Fields That Shadow Record Methods

```go
// BAD - Creates Email() method that shadows Record.Email()
type User struct {
    core.BaseRecordProxy
}

func (u *User) Email() string {  // SHADOW!
    return u.GetString("email")
}
```

### 7.3 DON'T: Rename to Avoid Shadowing (Usually Wrong)

While technically possible:

```go
// NOT RECOMMENDED - Technically works but defeats the purpose
type User struct {
    core.BaseRecordProxy
}

// schema-name: email
func (u *User) UserEmail() string {
    return u.GetString("email")
}
```

This is a code smell that indicates your schema design is conflicting with PocketBase's framework.

### 7.4 DO: Use "// system:" Comments for System Fields

If you need to customize system fields, mark them clearly:

```go
type User struct {
    core.BaseRecordProxy
}

// system: email
// Custom email getter with extra logic
func (u *User) GetEmail() string {
    email := u.Email()  // Calls Record.Email()
    // ... do something special
    return email
}
```

---

## 8. Technical Deep Dive: Why Shadowing Breaks the Framework

### 8.1 Method Promotion in Go

When you embed a pointer in a struct, all its methods are promoted:

```go
type User struct {
    *core.Record  // or via BaseRecordProxy
}

user := &User{}
user.Email()  // This calls Record.Email() via promotion
```

When you define a method with the same name on the outer struct, it shadows the promoted method:

```go
type User struct {
    core.BaseRecordProxy
}

func (u *User) Email() string {
    // This shadows the promoted Record.Email()
    return u.GetString("email")
}

user := &User{}
user.Email()  // Calls the defined method, not Record.Email()
```

### 8.2 Interface Type Assertions Break

If code asserts the proxy to the RecordProxy interface and expects Record methods:

```go
var p core.RecordProxy = &User{}
record := p.ProxyRecord()
record.Email()  // Works - calls Record.Email()

// But if User shadows Email(), you get unexpected behavior
```

### 8.3 Framework Internals Assume These Methods Exist

From `record_model.go` line 1438-1441:

```go
if lastSavedRecord.TokenKey() == e.Record.TokenKey() &&
    (lastSavedRecord.Get(FieldNamePassword) != e.Record.Get(FieldNamePassword) ||
        lastSavedRecord.Email() != e.Record.Email()) {
    e.Record.RefreshTokenKey()
}
```

The framework calls these methods internally. If a proxy shadows them with incompatible behavior, the framework breaks.

---

## 9. Summary Table

| Item | What It Is | Why It Matters |
|------|-----------|----------------|
| `core.Record` | Base model for all data records | Framework foundation |
| `BaseRecordProxy` | Interface for typed proxies | Template for generating proxy structs |
| Embedded `*Record` | Pointer field in proxy | Allows method promotion |
| `Email()`, `TokenKey()` | Auth framework methods | Special behavior for auth collections |
| Field names like `email` | Schema field from PocketBase | Stored in Record.data |
| Generated getters/setters | Code produced from template | Call Record.Get/Set under the hood |
| **Shadowing** | Generated method = Record method name | **BREAKS FRAMEWORK INTERNALS** |

---

## 10. Conclusion

### The Real Issue

The shadowing problem exists because:

1. PocketBase provides convenience methods on `Record` for auth fields (`Email()`, `TokenKey()`, etc.)
2. Generated proxy code creates getters with the same names for database fields
3. These conflicts prevent the framework from safely working with your proxies
4. The code generator **detects this at build time and refuses to compile** to prevent subtle bugs

### The Solution

**DO NOT rename fields to work around shadowing.** Instead:

1. For standard fields (title, name, etc.) - create proxy getters/setters freely
2. For system fields (email, password, tokenKey) in auth collections - rely on the inherited Record methods
3. If you need custom behavior around system fields, write wrapper methods with different names

### Key Takeaway

The fact that PocketBase's code generator explicitly prevents shadowing (with a build error) shows this is a **serious design issue**, not a minor conflict. Trying to work around it by renaming fields is fighting the framework, not working with it.

The framework is telling you: "These methods must not be overridden, or I can't work correctly." Listen to it.
