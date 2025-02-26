package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dangerclosesec/supra/permissions/migration"
	"github.com/dangerclosesec/supra/permissions/parser"
	"github.com/spf13/cobra"
)

var (
	dbConnString string
	verbose      bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&dbConnString, "db", "d", "", "Database connection string")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(versionCmd)
}

var rootCmd = &cobra.Command{
	Use:   "permify",
	Short: "Permify is a CLI tool for managing permission models",
	Long:  `Permify is a CLI tool for parsing, migrating, and versioning permission models.`,
}

var parseCmd = &cobra.Command{
	Use:   "parse [file]",
	Short: "Parse a .perm file",
	Long:  `Parse a .perm file and display its contents.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		model, errors, err := parser.ParseFile(filePath)
		if err != nil {
			log.Fatalf("Failed to parse file: %v", err)
		}

		if len(errors) > 0 {
			fmt.Println("Parsing errors:")
			for _, err := range errors {
				fmt.Println("  - " + err)
			}
			os.Exit(1)
		}

		fmt.Printf("Successfully parsed %s\n", filePath)
		fmt.Printf("Found %d entities\n", len(model.Entities))

		if verbose {
			for name, entity := range model.Entities {
				fmt.Printf("\nEntity: %s\n", name)

				fmt.Printf("  Relations (%d):\n", len(entity.Relations))
				for _, rel := range entity.Relations {
					fmt.Printf("    - %s @%s\n", rel.Name, rel.Target)
				}

				fmt.Printf("  Permissions (%d):\n", len(entity.Permissions))
				for _, perm := range entity.Permissions {
					fmt.Printf("    - %s = %s\n", perm.Name, perm.Expression)
				}
			}
		}
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database schema",
	Long:  `Initialize the database schema for permission models.`,
	Run: func(cmd *cobra.Command, args []string) {
		if dbConnString == "" {
			log.Fatal("Database connection string is required")
		}

		db, err := sql.Open("postgres", dbConnString)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		migrator := migration.NewMigrator(db)
		err = migrator.InitializeSchema()
		if err != nil {
			log.Fatalf("Failed to initialize schema: %v", err)
		}

		fmt.Println("Schema initialized successfully")
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate [file]",
	Short: "Apply a permission model to the database",
	Long:  `Parse a .perm file and apply it to the database.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		if dbConnString == "" {
			log.Fatal("Database connection string is required")
		}

		db, err := sql.Open("postgres", dbConnString)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		model, errors, err := parser.ParseFile(filePath)
		if err != nil {
			log.Fatalf("Failed to parse file: %v", err)
		}

		if len(errors) > 0 {
			fmt.Println("Parsing errors:")
			for _, err := range errors {
				fmt.Println("  - " + err)
			}
			os.Exit(1)
		}

		migrator := migration.NewMigrator(db)

		// Initialize schema if needed
		err = migrator.InitializeSchema()
		if err != nil {
			log.Fatalf("Failed to initialize schema: %v", err)
		}

		// Generate description
		description := fmt.Sprintf("Migration from %s at %s",
			filepath.Base(filePath), time.Now().Format(time.RFC3339))

		// Apply migration
		diff, err := migrator.ApplyMigration(model, description)
		if err != nil {
			log.Fatalf("Failed to apply migration: %v", err)
		}

		if diff == "No changes detected. Migration skipped." {
			fmt.Println(diff)
			return
		}

		fmt.Println("Migration applied successfully")

		if verbose {
			fmt.Println("\nChanges:")
			fmt.Println(diff)
		}

		// Get current version
		version, err := migrator.GetCurrentVersion()
		if err != nil {
			log.Fatalf("Failed to get current version: %v", err)
		}

		fmt.Printf("Current version: %d\n", version)
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff [file]",
	Short: "Show differences between a .perm file and the current database",
	Long:  `Parse a .perm file and show differences compared to the current database.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		if dbConnString == "" {
			log.Fatal("Database connection string is required")
		}

		db, err := sql.Open("postgres", dbConnString)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		// Parse the file
		model, errors, err := parser.ParseFile(filePath)
		if err != nil {
			log.Fatalf("Failed to parse file: %v", err)
		}

		if len(errors) > 0 {
			fmt.Println("Parsing errors:")
			for _, err := range errors {
				fmt.Println("  - " + err)
			}
			os.Exit(1)
		}

		// Initialize migrator
		migrator := migration.NewMigrator(db)

		// Load current model
		currentModel, err := migrator.LoadCurrentModel()
		if err != nil {
			log.Fatalf("Failed to load current model: %v", err)
		}

		// Generate diff
		diff := migration.GenerateDiff(currentModel, model)

		// Print diff
		fmt.Println(diff.String())
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the current permission model version",
	Long:  `Show the current permission model version in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		if dbConnString == "" {
			log.Fatal("Database connection string is required")
		}

		db, err := sql.Open("postgres", dbConnString)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		migrator := migration.NewMigrator(db)
		version, err := migrator.GetCurrentVersion()
		if err != nil {
			log.Fatalf("Failed to get current version: %v", err)
		}

		fmt.Printf("Current permission model version: %d\n", version)

		if verbose {
			// Get version history
			rows, err := db.Query(`
				SELECT version, description, source_file, applied_at 
				FROM permission_versions 
				ORDER BY version DESC
			`)
			if err != nil {
				log.Fatalf("Failed to get version history: %v", err)
			}
			defer rows.Close()

			fmt.Println("\nVersion history:")
			fmt.Println("----------------")

			for rows.Next() {
				var v int
				var desc, sourceFile string
				var appliedAt time.Time
				if err := rows.Scan(&v, &desc, &sourceFile, &appliedAt); err != nil {
					log.Fatalf("Failed to scan version: %v", err)
				}

				fmt.Printf("Version %d (applied %s)\n", v, appliedAt.Format(time.RFC3339))
				fmt.Printf("  Source: %s\n", sourceFile)
				fmt.Printf("  Description: %s\n\n", desc)
			}
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
