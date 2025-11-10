package cli

import (
	"fmt"
	"os"
	"path/filepath"

	pdfform "github.com/joeblew999/wellknown/pkg/pdf"
	"github.com/joeblew999/wellknown/pkg/pdf/web"
	"github.com/spf13/cobra"
)

// Run executes the CLI with the provided configuration
func Run(cfg *pdfform.Config) error {
	rootCmd := &cobra.Command{
		Use:   "pdfform",
		Short: "Fill PDF forms with a simple 5-step workflow",
		Long: `pdfform - Fill PDF forms with JSON data

📋 WORKFLOW - Follow these numbered steps:

1️⃣  BROWSE FORMS - Find Australian government forms
    pdfform 1-browse                    # List all states
    pdfform 1-browse --state VIC        # Show Victoria forms

2️⃣  DOWNLOAD FORM - Get the PDF you need
    pdfform 2-download F3520            # Download by form code
    pdfform 2-download F3520 -o pdfs/   # Save to directory

3️⃣  INSPECT FIELDS - See what fields the form has
    pdfform 3-inspect form.pdf          # Creates template JSON
    pdfform 3-inspect form.pdf -o out.json

4️⃣  FILL FORM - Fill in your data
    pdfform 4-fill data.json            # Fill the form
    pdfform 4-fill data.json --flatten  # Fill and lock
    pdfform 4-fill --test vba_basic     # Use test case

5️⃣  TEST - Run automated tests (optional)
    pdfform 5-test                      # List all tests
    pdfform 5-test vba_basic            # Run specific test
    pdfform 5-test --all                # Run all tests

Each step guides you to the next! Just follow the numbers.`,
	}

	// ========================================
	// 1️⃣ BROWSE FORMS
	// ========================================
	var browseState string
	browseCmd := &cobra.Command{
		Use:   "1-browse",
		Short: "1️⃣  Browse available Australian government forms",
		Long: `Step 1: Browse the catalog of government forms

Examples:
  pdfform 1-browse               # List all forms
  pdfform 1-browse --state VIC   # List Victoria forms
  pdfform 1-browse --state NSW   # List NSW forms`,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := pdfform.Browse(pdfform.BrowseOptions{
				CatalogPath: cfg.CatalogFilePath(),
				State:       browseState,
			})
			if err != nil {
				return err
			}

			fmt.Println("1️⃣  BROWSE FORMS")
			fmt.Println()

			if browseState == "" {
				// Display all states
				fmt.Println("📍 Available States:")
				catalog, _ := pdfform.LoadFormsCatalog(cfg.CatalogFilePath())
				for _, state := range result.States {
					forms := catalog.GetFormsByState(state)
					fmt.Printf("   %s (%d form(s))\n", state, len(forms))
				}
				fmt.Println()
				fmt.Println("💡 Tip: Use --state to see forms for a specific state")
				fmt.Println("   Example: pdfform 1-browse --state VIC")
			} else {
				// Display forms for specific state
				fmt.Printf("📋 Forms for %s:\n\n", browseState)
				for i, form := range result.Forms {
					fmt.Printf("%d. %s", i+1, form.FormName)
					if form.FormCode != "" {
						fmt.Printf(" (Code: %s)", form.FormCode)
					}
					fmt.Println()
					fmt.Printf("   Format: %s", form.Format)
					if form.OnlineAvailable {
						fmt.Print(" | Online: Yes")
					}
					fmt.Println()
					if form.Notes != "" {
						fmt.Printf("   Notes: %s\n", form.Notes)
					}
					fmt.Println()
				}

				fmt.Println("➡️  Next Step: Download a form")
				if len(result.Forms) > 0 && result.Forms[0].FormCode != "" {
					fmt.Printf("   pdfform 2-download %s\n", result.Forms[0].FormCode)
				}
			}

			return nil
		},
	}
	browseCmd.Flags().StringVarP(&browseState, "state", "s", "", "Filter by state (VIC, NSW, QLD, SA, WA, TAS, ACT, NT)")

	// ========================================
	// 2️⃣ DOWNLOAD FORM
	// ========================================
	var downloadOutDir string
	downloadCmd := &cobra.Command{
		Use:   "2-download [form-code]",
		Short: "2️⃣  Download a government form by its code",
		Long: `Step 2: Download a PDF form

Examples:
  pdfform 2-download F3520              # Download Queensland form
  pdfform 2-download VRPIN00613         # Download Victoria form
  pdfform 2-download F3520 -o pdfs/     # Download to specific directory`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			formCode := args[0]

			if downloadOutDir == "" {
				downloadOutDir = cfg.DownloadsPath()
			}

			fmt.Println("2️⃣  DOWNLOAD FORM")
			fmt.Println()

			result, err := pdfform.Download(pdfform.DownloadOptions{
				CatalogPath: cfg.CatalogFilePath(),
				FormCode:    formCode,
				OutputDir:   downloadOutDir,
			})
			if err != nil {
				if err.Error() == fmt.Sprintf("form with code '%s' not found", formCode) {
					return fmt.Errorf("%w\n\n💡 Tip: Use 'pdfform 1-browse --state VIC' to see available forms", err)
				}
				return err
			}

			fmt.Printf("📥 Downloaded: %s (%s)\n", result.Form.FormName, result.Form.FormCode)
			fmt.Printf("✅ Downloaded to: %s\n\n", result.PDFPath)

			fmt.Println("➡️  Next Step: Inspect the form fields")
			fmt.Printf("   pdfform 3-inspect %s\n", result.PDFPath)

			return nil
		},
	}
	downloadCmd.Flags().StringVarP(&downloadOutDir, "output-dir", "o", "", "Output directory for downloaded form (default: data/downloads)")

	// ========================================
	// 3️⃣ INSPECT FIELDS
	// ========================================
	var inspectOut string
	inspectStepCmd := &cobra.Command{
		Use:   "3-inspect [pdf-file]",
		Short: "3️⃣  Extract form fields from a PDF to create JSON template",
		Long: `Step 3: Inspect a PDF and extract all fillable fields

Examples:
  pdfform 3-inspect form.pdf                    # Creates form_fields.json
  pdfform 3-inspect form.pdf -o template.json   # Custom output name`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pdfFile := args[0]

			outputDir := inspectOut
			if outputDir == "" {
				outputDir = cfg.TemplatesPath()
			}

			fmt.Println("3️⃣  INSPECT FIELDS")
			fmt.Println()

			result, err := pdfform.Inspect(pdfform.InspectOptions{
				PDFPath:   pdfFile,
				OutputDir: outputDir,
			})
			if err != nil {
				return err
			}

			fmt.Printf("🔍 Inspecting: %s\n", filepath.Base(pdfFile))
			fmt.Printf("✅ Found %d form fields\n", result.FieldCount)
			fmt.Printf("✅ Template saved to: %s\n\n", result.TemplatePath)

			fmt.Println("➡️  Next Step: Edit the template and fill in your data")
			fmt.Printf("   1. Open %s in your editor\n", result.TemplatePath)
			fmt.Println("   2. Fill in the field values")
			fmt.Println("   3. Run: pdfform 4-fill " + result.TemplatePath)

			return nil
		},
	}
	inspectStepCmd.Flags().StringVarP(&inspectOut, "output", "o", "", "Output directory or file (default: data/templates/<pdfname>_template.json)")

	// ========================================
	// 4️⃣ FILL FORM
	// ========================================
	var fillFlatten bool
	var fillOutput string
	var fillTest string
	fillStepCmd := &cobra.Command{
		Use:   "4-fill [data.json]",
		Short: "4️⃣  Fill a PDF form with your data",
		Long: `Step 4: Fill a PDF form using JSON data

The JSON file should contain:
  - pdf_url: Path or URL to the PDF form
  - fields: Object with field names and values

Examples:
  pdfform 4-fill data.json                # Fill form
  pdfform 4-fill data.json --flatten      # Fill and lock fields
  pdfform 4-fill data.json -o output.pdf  # Custom output name
  pdfform 4-fill --test vba_basic         # Fill using test case`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var dataFile string

			// Handle --test flag
			if fillTest != "" {
				dataFile = filepath.Join(cfg.TestScenariosPath(), fillTest+".json")
				if fillOutput == "" {
					fillOutput = filepath.Join(cfg.OutputsPath(), fillTest+"_filled.pdf")
				}
			} else {
				if len(args) == 0 {
					return fmt.Errorf("data.json file required (or use --test flag)")
				}
				dataFile = args[0]
			}

			if fillOutput == "" {
				fillOutput = cfg.OutputsPath()
			}

			fmt.Println("4️⃣  FILL FORM")
			fmt.Println()

			result, err := pdfform.Fill(pdfform.FillOptions{
				DataPath:  dataFile,
				OutputDir: fillOutput,
				Flatten:   fillFlatten,
			})
			if err != nil {
				return err
			}

			fmt.Printf("📥 Processing: %s\n", filepath.Base(dataFile))
			if fillFlatten {
				fmt.Println("🔒 Flattening PDF (locking fields)...")
				fmt.Printf("✅ Flattened PDF: %s\n", result.OutputPath)
			} else {
				fmt.Printf("✅ Filled PDF: %s\n", result.OutputPath)
			}

			fmt.Println()
			fmt.Printf("🎉 SUCCESS! Your form is ready: %s\n", result.OutputPath)
			if result.InputPDF != "" {
				fmt.Printf("   Original form: %s\n", result.InputPDF)
			}

			return nil
		},
	}
	fillStepCmd.Flags().BoolVar(&fillFlatten, "flatten", false, "Lock form fields (make read-only)")
	fillStepCmd.Flags().StringVarP(&fillOutput, "output", "o", "", "Output directory or file (default: data/outputs/<datafile>_filled.pdf)")
	fillStepCmd.Flags().StringVarP(&fillTest, "test", "t", "", "Load test case from data/cases/test_scenarios/<name>.json")

	// ========================================
	// 5️⃣ TEST
	// ========================================
	testStepCmd := &cobra.Command{
		Use:   "5-test [test-name]",
		Short: "5️⃣  Run automated tests on PDF forms",
		Long: `Step 5: Run test cases to verify form filling works correctly

Examples:
  pdfform 5-test              # List all test cases
  pdfform 5-test vba_basic    # Run specific test case
  pdfform 5-test --all        # Run all test cases`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runAll, _ := cmd.Flags().GetBool("all")

			fmt.Println("5️⃣  TEST")
			fmt.Println()

			if len(args) == 0 && !runAll {
				// List available test cases
				testCasesDir := cfg.TestScenariosPath()
				cases, err := pdfform.ListTestCases(testCasesDir)
				if err != nil {
					return fmt.Errorf("failed to list test cases: %w", err)
				}
				if len(cases) == 0 {
					fmt.Printf("No test cases found in %s\n", testCasesDir)
					fmt.Println()
					fmt.Println("💡 Tip: Create test cases in data/cases/test_scenarios/ to automate form filling")
					return nil
				}
				fmt.Println("📋 Available test cases:")
				for i, c := range cases {
					name := filepath.Base(c)
					name = name[:len(name)-len(filepath.Ext(name))]
					fmt.Printf("   %d. %s\n", i+1, name)
				}
				fmt.Println()
				fmt.Println("➡️  Run a test:")
				if len(cases) > 0 {
					name := filepath.Base(cases[0])
					name = name[:len(name)-len(filepath.Ext(name))]
					fmt.Printf("   pdfform 5-test %s\n", name)
				}
				return nil
			}

			// Run specific test or all tests
			var testNames []string
			testCasesDir := cfg.TestScenariosPath()
			if runAll {
				cases, err := pdfform.ListTestCases(testCasesDir)
				if err != nil {
					return err
				}
				for _, c := range cases {
					name := filepath.Base(c)
					name = name[:len(name)-len(filepath.Ext(name))]
					testNames = append(testNames, name)
				}
			} else {
				testNames = args
			}

			// Ensure output directory exists
			outputDir := cfg.OutputsPath()
			os.MkdirAll(outputDir, 0755)

			passed := 0
			failed := 0
			for _, name := range testNames {
				testFile := filepath.Join(testCasesDir, name+".json")
				fmt.Printf("🧪 Running: %s\n", name)

				result, err := pdfform.Test(pdfform.TestOptions{
					TestCasePath: testFile,
					OutputDir:    outputDir,
				})
				if err != nil || !result.Passed {
					fmt.Printf("   ❌ Failed: %v\n\n", result.Error)
					failed++
					continue
				}
				fmt.Printf("   ✅ Passed: %s\n\n", result.OutputPath)
				passed++
			}

			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			fmt.Printf("Results: %d passed, %d failed\n", passed, failed)
			if failed == 0 {
				fmt.Println("🎉 All tests passed!")
			}

			return nil
		},
	}
	testStepCmd.Flags().Bool("all", false, "Run all test cases")

	// ========================================
	// SERVE - Web Server
	// ========================================
	var servePort int
	var serveHTTP bool
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "🌐 Start web server with 5-step workflow GUI",
		Long: `Start web server for PDF form filling

The web interface mirrors the 5-step CLI workflow:
  https://localhost:8080/              → Home
  https://localhost:8080/1-browse      → Browse forms
  https://localhost:8080/2-download    → Download form
  https://localhost:8080/3-inspect     → Inspect fields
  https://localhost:8080/4-fill        → Fill form
  https://localhost:8080/5-test        → Run tests

HTTPS Support:
  By default, the server uses HTTPS for mobile device support.
  Certificates are auto-generated using mkcert and stored in .data/certs/

Examples:
  pdfform serve                       # Start HTTPS on port 8080
  pdfform serve --port 3000           # Start on custom port
  pdfform serve --http                # Force HTTP (not recommended for mobile)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to HTTPS unless --http is specified
			useHTTPS := !serveHTTP

			if useHTTPS {
				// Ensure certificates exist
				cm := pdfform.NewCertManager(cfg)
				if err := cm.EnsureCerts(); err != nil {
					fmt.Printf("⚠️  Failed to setup HTTPS: %v\n", err)
					fmt.Println("   Falling back to HTTP...")
					useHTTPS = false
				}
			}

			// Show connection info
			pdfform.PrintServerInfo(fmt.Sprintf("%d", servePort), useHTTPS)

			return web.StartServer(servePort, cfg, useHTTPS)
		},
	}
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to run the server on")
	serveCmd.Flags().BoolVar(&serveHTTP, "http", false, "Use HTTP instead of HTTPS (not recommended for mobile)")

	// ========================================
	// CERTS - Certificate Management
	// ========================================
	certsCmd := &cobra.Command{
		Use:   "certs",
		Short: "🔒 Manage HTTPS certificates",
		Long: `Manage HTTPS certificates for local development

Subcommands:
  pdfform certs info       # Show certificate information
  pdfform certs generate   # Generate new certificates
  pdfform certs regenerate # Regenerate existing certificates`,
	}

	certsInfoCmd := &cobra.Command{
		Use:   "info",
		Short: "Show certificate information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cm := pdfform.NewCertManager(cfg)
			return cm.ShowCertInfo()
		},
	}

	certsGenerateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate new certificates",
		RunE: func(cmd *cobra.Command, args []string) error {
			cm := pdfform.NewCertManager(cfg)
			return cm.EnsureCerts()
		},
	}

	certsRegenerateCmd := &cobra.Command{
		Use:   "regenerate",
		Short: "Regenerate existing certificates",
		RunE: func(cmd *cobra.Command, args []string) error {
			cm := pdfform.NewCertManager(cfg)
			return cm.RegenerateCerts()
		},
	}

	certsCmd.AddCommand(certsInfoCmd)
	certsCmd.AddCommand(certsGenerateCmd)
	certsCmd.AddCommand(certsRegenerateCmd)

	// Add numbered workflow commands
	rootCmd.AddCommand(browseCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(inspectStepCmd)
	rootCmd.AddCommand(fillStepCmd)
	rootCmd.AddCommand(testStepCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(certsCmd)

	// Show help by default if no command specified
	validCommands := map[string]bool{
		"1-browse":   true,
		"2-download": true,
		"3-inspect":  true,
		"4-fill":     true,
		"5-test":     true,
		"serve":      true,
		"certs":      true,
		"help":       true,
		"--help":     true,
		"-h":         true,
		"completion": true,
	}

	if len(os.Args) == 1 || (len(os.Args) > 1 && !validCommands[os.Args[1]]) {
		// Show help if no command or unknown command
		os.Args = append([]string{os.Args[0], "--help"}, os.Args[1:]...)
	}

	return rootCmd.Execute()
}
