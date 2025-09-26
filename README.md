# Go CSV Merger Utility

A simple and robust command-line utility written in Go to merge multiple CSV files from an input directory into a single output file. It is designed with automation in mind, providing structured logging, file archiving, and cleanup upon successful completion.

## Features ✨

* **Merge CSVs**: Combines all `.csv` files from a source directory.
* **Dynamic Filename Column**: Automatically adds a `tick_nm` column to the merged file, populated with the source filename (without the extension).
* **Structured Logging**: Creates a timestamped log file for each run, recording all actions and errors.
* **Archiving**: On successful completion, archives all source files and the final merged file to a timestamped folder.
* **Automatic Cleanup**: Source and output directories are cleaned up after a successful run by moving files to the archive.
* **Cross-Platform**: Built to run on macOS, Linux, and Windows.

## Prerequisites

* You must have **Go version 1.16 or newer** installed.

## How to Use

1.  Clone the repository:
    ```bash
    git clone [https://github.com/JarredKF/go-csv-merger.git](https://github.com/JarredKF/go-csv-merger.git)
    cd go-csv-merger
    ```

2.  Run the program using `go run`, providing the four required directory paths as flags.

    **Example Command:**
    ```bash
    go run main.go \
      -datin="./path/to/input_files" \
      -datout="./path/to/output_dir" \
      -datlog="./path/to/log_dir" \
      -arch="./path/to/archive_dir"
    ```

### Command-Line Flags

* `-datin`: (Required) Path to the input directory containing the source `.csv` files.
* `-datout`: (Required) Path to the output directory where the merged file will be written.
* `-datlog`: (Required) Path to a directory where the timestamped log file will be stored.
* `-arch`: (Required) Path to a directory where all source files and the final merged file will be archived upon success.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
