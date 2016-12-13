package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

var (
	Commit string
	Tag    string
)

func check(zipDirectoryPath string, serialNumbers map[string]bool) ([]string, error) {
	missing := make([]string, 0)
	fileInfos, err := ioutil.ReadDir(zipDirectoryPath)
	if err != nil {
		log.Printf("failed reading zip directory. %s", err)
		return missing, err
	}
	found := false
	for _, fileInfo := range fileInfos {
		if !strings.HasSuffix(fileInfo.Name(), "zip") {
			continue
		}

		filePath := path.Join(zipDirectoryPath, fileInfo.Name())

		reader, err := zip.OpenReader(filePath)
		if err != nil {
			return missing, err
		}
		found = true
		for _, file := range reader.File {
			fname := file.FileInfo().Name()
			if !strings.HasPrefix(fname, "9050000") {
				continue
			}
			_, found := serialNumbers[fname]
			if found {
				delete(serialNumbers, fname)
			}

		}
	}

	if !found {
		return missing, errors.New("No zip files found")
	}
	if len(serialNumbers) > 0 {
		for k, _ := range serialNumbers {
			missing = append(missing, k)
		}

		return missing, errors.New("Missing SNs")
	}

	return missing, nil
}

func main() {
	fmt.Println("Tool: Jabil Pill Checker")
	fmt.Printf("Version: %s [%s]\n", Tag, Commit)

	if len(os.Args) != 3 {
		fmt.Printf("Usage: pill-checker-%s.exe {excel_file}.xlsx path/to/zip/directory/\n", Tag)
		os.Exit(0)
	}
	excelFileName := os.Args[1]
	zipDirectory := os.Args[2]

	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		log.Fatal(err)
	}

	serialNumbers := make(map[string]bool)

	for _, sheet := range xlFile.Sheets {

		for i, row := range sheet.Rows {
			if i == 0 {
				header, err := row.Cells[3].String()
				if err != nil {
					log.Fatal(err)
				}

				trimmedHeader := strings.TrimSpace(header)
				if trimmedHeader != "PCBASerialNo" {
					log.Fatal("header doesn't match: PCBASerialNo -> " + trimmedHeader)
				}
				continue
			}

			sn, err := row.Cells[3].String()
			if err != nil {
				log.Fatal(err)
			}
			if !strings.HasPrefix(sn, "9050") {
				log.Fatal(fmt.Sprintf("Bad prefix row: %d. -> %s", i+1, sn))
			}
			serialNumbers[sn] = true
		}
	}

	fmt.Println("SNs to check:", len(serialNumbers))
	missing, snErr := check(zipDirectory, serialNumbers)
	fmt.Println("SNs not found:", len(serialNumbers))
	if snErr != nil {
		fmt.Println(snErr)
		if len(missing) > 0 {
			content := strings.Join(missing, "\n")
			now := time.Now()
			outputFilename := fmt.Sprintf("missing-%s.txt", now.Format("2006-01-02T15-04-05"))
			ioutil.WriteFile(outputFilename, []byte(content), 0644)
			fmt.Printf("Output: %s\n", outputFilename)
		}
		os.Exit(1)
	}
	fmt.Println("-> OK to send to Hello.")
}
