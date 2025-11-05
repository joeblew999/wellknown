/// <reference path="../pb_data/types.d.ts" />

// Example PocketBase JS hook
// This demonstrates the jsvm plugin functionality
// See: https://pocketbase.io/docs/js-overview/

onRecordCreateRequest((e) => {
    console.log(`📝 New record created in collection: ${e.record.collection().name}`)
    return e.next()
}, "google_tokens")

onRecordAfterCreateSuccess((e) => {
    console.log(`✅ Google token stored for user: ${e.record.getString("user_id")}`)
    return e.next()
}, "google_tokens")
