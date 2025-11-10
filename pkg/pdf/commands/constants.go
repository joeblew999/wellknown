package commands

import "os"

// File suffixes used across commands
const (
	FilledPDFSuffix    = "_filled.pdf"
	FlatPDFSuffix      = "_flat.pdf"
	TemplateJSONSuffix = "_template.json"
	MetaJSONSuffix     = ".meta.json"
)

// Download progress stages
const (
	DownloadStageFoundForm     = "found_form"
	DownloadStageDownloading   = "downloading"
	DownloadStageSavingMeta    = "saving_metadata"
	DownloadStageCreateDir     = "create_dir"
	DownloadStageDownloadPDF   = "download_pdf"
)

// Generic operation stages
const (
	StageCreateDir  = "create_dir"
	StageListFields = "list_fields"
	StageExportJSON = "export_json"
	StageFillPDF    = "fill_pdf"
	StageFlatten    = "flatten"
	StageLoadCase   = "load_case"
	StageSaveCase   = "save_case"
	StageCreate     = "create"
	StageLoad       = "load"
	StageSave       = "save"
)

// Progress values for download operations
const (
	ProgressFoundForm   = 0.2
	ProgressDownloading = 0.4
	ProgressMetadata    = 0.8
	ProgressComplete    = 1.0
)

// File permissions
const (
	DefaultDirPerm os.FileMode = 0755
)
