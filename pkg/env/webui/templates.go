package webui

// Pico CSS CDN link - classless CSS framework with dark mode support
const picoCSSLink = `<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">`

// Minimal custom styles optimized for developer workflow
const customStyles = `
<style>
/* Compact header */
header { padding: 1rem 0; border-bottom: 1px solid var(--pico-muted-border-color); }
header h2 { margin: 0; }
.stats { display: flex; gap: 1rem; font-size: 0.9rem; color: var(--pico-muted-color); }
.stats span { display: flex; align-items: center; gap: 0.25rem; }

/* Filter box */
#filter { margin: 1rem 0; }

/* Compact table */
table { font-size: 0.9rem; }
table td, table th { padding: 0.5rem; vertical-align: top; }
table th { font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.5px; }

/* Variable name - bold monospace */
.var-name {
    font-family: monospace;
    font-weight: 600;
    font-size: 0.95rem;
}

/* Value display with copy button */
.value-cell {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-family: monospace;
    font-size: 0.9rem;
}
.copy-btn {
    opacity: 0.3;
    cursor: pointer;
    border: none;
    background: none;
    padding: 0.25rem;
    font-size: 1rem;
}
.copy-btn:hover { opacity: 1; }
tr:hover .copy-btn { opacity: 0.6; }

/* Status badges - minimal */
.status {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-right: 0.5rem;
}
.status-set { background: #28a745; }
.status-missing { background: #ffc107; }
.status-empty { background: #6c757d; opacity: 0.3; }

/* Tags - ultra minimal */
.tag {
    font-size: 0.7rem;
    padding: 0.1rem 0.4rem;
    border-radius: 2px;
    font-weight: 500;
    margin-right: 0.25rem;
    opacity: 0.7;
}
.tag-secret { background: #dc3545; color: white; }
.tag-required { background: #ffc107; color: #000; }

/* Secret values */
.secret { color: #dc3545; font-weight: 600; }
.empty { color: var(--pico-muted-color); font-style: italic; }

/* Export buttons */
.export-bar {
    display: flex;
    gap: 0.5rem;
    margin: 1rem 0;
    flex-wrap: wrap;
}
.export-bar button {
    font-size: 0.8rem;
    padding: 0.4rem 0.8rem;
}

/* Highlight missing required vars */
tr.missing-required { background: rgba(255, 193, 7, 0.1); }

/* Hidden rows (for filter) */
tr.hidden { display: none; }
</style>
`
