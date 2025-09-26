package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Use flags to accept command-line arguments.
	datinDir := flag.String("datin", "", "Input directory for ticker CSV files (required)")
	datoutDir := flag.String("datout", "", "Output directory for the merged file (required)")
	datlogDir := flag.String("datlog", "", "Directory for log files (required)")
	archDir := flag.String("arch", "", "Directory to archive source and merged files (required)")
	flag.Parse()

	// Validate that all required arguments are provided.
	if *datinDir == "" || *datoutDir == "" || *datlogDir == "" || *archDir == "" {
		fmt.Println("Error: All flags (-datin, -datout, -datlog, -arch) are required.")
		flag.Usage()
		os.Exit(1)
	}

	// Set up structured logging to a file.
	// This function is the key to creating the log file.
	if err := setupLogger(*datlogDir); err != nil {
		// If logger fails, we can't log, so we panic.
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	log.Println("Process started.")
	log.Printf("Input Directory: %s", *datinDir)
	log.Printf("Output Directory: %s", *datoutDir)
	log.Printf("Log Directory: %s", *datlogDir)
	log.Printf("Archive Directory: %s", *archDir)

	// Core logic is wrapped to handle errors gracefully.
	outputFilePath, err := processFiles(*datinDir, *datoutDir)
	if err != nil {
		log.Fatalf("FATAL: File processing failed: %v", err)
	}

	// On success, run the archiving and cleanup process.
	if err := archiveAndCleanup(*archDir, outputFilePath, *datinDir); err != nil {
		log.Fatalf("FATAL: Archiving and cleanup failed: %v", err)
	}

	log.Println("Process completed successfully.")
}

// setupLogger configures the log package to write to a timestamped file in the log directory.
func setupLogger(logDir string) error {
	// Create the log directory if it doesn't exist.
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("could not create log directory %s: %w", logDir, err)
	}

	// Create a unique, timestamped log file name.
	logFile := filepath.Join(logDir, fmt.Sprintf("merge_process_%s.log", time.Now().Format("20060102_150405")))
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		return fmt.Errorf("could not open log file %s: %w", logFile, err)
	}

	// Use io.MultiWriter to send log output to BOTH the console and the file.
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
	return nil
}

// processFiles merges all CSVs from datinDir into a single file in datoutDir.
func processFiles(datinDir, datoutDir string) (string, error) {
	if err := os.MkdirAll(datoutDir, 0755); err != nil {
		return "", fmt.Errorf("could not create output directory %s: %w", datoutDir, err)
	}

	dateStr := time.Now().Format("20060102")
	outputFile := filepath.Join(datoutDir, fmt.Sprintf("extract_%s.csv", dateStr))
	outF, err := os.Create(outputFile)
	if err != nil {
		return "", fmt.Errorf("could not create output file %s: %w", outputFile, err)
	}
	defer outF.Close()

	writer := csv.NewWriter(outF)
	defer writer.Flush()
	headerWritten := false
	filesProcessed := 0

	log.Println("Starting to walk input directory...")
	err = filepath.Walk(datinDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil // Skip directories
		}

		if strings.HasSuffix(strings.ToLower(info.Name()), ".csv") {
			// This is the line that logs each file as it's being processed.
			log.Printf("Processing file: %s", info.Name())
			ticker := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))

			inF, err := os.Open(path)
			if err != nil {
				log.Printf("WARNING: Could not open file %s, skipping. Error: %v", path, err)
				return nil
			}
			defer inF.Close()

			reader := csv.NewReader(inF)
			records, err := reader.ReadAll()
			if err != nil {
				log.Printf("WARNING: Could not read CSV data from %s, skipping. Error: %v", path, err)
				return nil
			}

			if len(records) < 2 {
				log.Printf("INFO: Skipping empty or header-only file: %s", info.Name())
				return nil
			}

			if !headerWritten {
				header := append(records[0], "tick_nm")
				if err := writer.Write(header); err != nil {
					return fmt.Errorf("failed to write header to output file: %w", err)
				}
				headerWritten = true
			}

			for i := 1; i < len(records); i++ {
				row := append(records[i], ticker)
				if err := writer.Write(row); err != nil {
					return fmt.Errorf("failed to write row to output file: %w", err)
				}
			}
			filesProcessed++
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error during directory walk: %w", err)
	}

	log.Printf("Finished processing. Merged %d files into %s", filesProcessed, outputFile)
	return outputFile, nil
}

// archiveAndCleanup moves the source files and the final merged file to a timestamped archive directory.
func archiveAndCleanup(archDir, mergedFilePath, datinDir string) error {
	archiveSubDir := filepath.Join(archDir, fmt.Sprintf("archive_%s", time.Now().Format("20060102_150405")))
	if err := os.MkdirAll(archiveSubDir, 0755); err != nil {
		return fmt.Errorf("could not create archive subdirectory %s: %w", archiveSubDir, err)
	}
	log.Printf("Created archive directory: %s", archiveSubDir)

	mergedFileName := filepath.Base(mergedFilePath)
	newMergedPath := filepath.Join(archiveSubDir, mergedFileName)
	log.Printf("Archiving merged file to %s", newMergedPath)
	if err := os.Rename(mergedFilePath, newMergedPath); err != nil {
		return fmt.Errorf("failed to archive merged file: %w", err)
	}

	log.Println("Archiving source files...")
	files, err := os.ReadDir(datinDir)
	if err != nil {
		return fmt.Errorf("could not read datin directory %s for archiving: %w", datinDir, err)
	}

	for _, file := range files {
		if !file.IsDir() {
			oldPath := filepath.Join(datinDir, file.Name())
			newPath := filepath.Join(archiveSubDir, file.Name())
			if err := os.Rename(oldPath, newPath); err != nil {
				log.Printf("WARNING: Failed to archive source file %s: %v", oldPath, err)
			}
		}
	}

	log.Println("Archiving and cleanup complete.")
	return nil
}
