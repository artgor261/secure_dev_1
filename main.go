package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/windows"

	"archive/zip"

	"github.com/shirou/gopsutil/v3/disk"
)

var freeBytesAvailable uint64
var totalNumberOfBytes uint64
var totalNumberOfFreeBytes uint64

var err error = windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr("C:"), &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes)

const (
	maxFilesInArchive int = 1000
	freePart float32 = 0.3
)

var	maxUncompressedSize float32 = freePart * float32(freeBytesAvailable) 


type Student struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Age     int    `json:"age"`
	Group   string `json:"group"`
}

type Person struct {
	XMLName xml.Name `xml:"person"`
	Name string `xml:"name"`
	Age int `xml:"age"`
}

func main() {
	for true {
		fmt.Println("\nInput: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if err != nil {
			fmt.Printf("Input error: %v", err)
		}

		switch command := input; command {

		case "fsinfo":
			FSInfo()

		case "create":
			CreateWrite(reader)

		case "del":
			DeleteFile(reader)

		case "student":
			var student Student
			FillStudent(&student, reader)
			CreateJSON(student, reader)
		
		case "add-zip":
			fmt.Println("Print zip archive name: ")
			zipPath, _ := reader.ReadString('\n')
			zipPath = strings.TrimSpace(zipPath)

			fmt.Println("\nPrint list of files: ")
			files, _ := reader.ReadString('\n')
			filesList := strings.Fields(files)

			AddZip(zipPath, filesList...)

		case "extract-zip":
			fmt.Println("\nPrint path of zip file: ")
			zipPath, _ := reader.ReadString('\n')
			zipPath = strings.TrimSpace(zipPath)

			fmt.Println("\nPrint destination directory: ")
			destDir, _ := reader.ReadString('\n')
			destDir = strings.TrimSpace(destDir)

			ExtractZip(zipPath, destDir)
		
		case "person":
			CreateXML(reader)
		
		default:
			fmt.Printf("'%s' does not exist\n", command)
		}
	}

}

func FSInfo() {
	partitions, err := disk.Partitions(true)
	if err != nil {
		fmt.Printf("Error during returning partitions: %v", err)
	}

	for _, p := range partitions {
		fmt.Printf("\nDevice: %s\nMountpoint: %s\nFstype: %s\n\n %s", p.Device, p.Mountpoint, p.Fstype, p.Opts)
	}
}

func CreateWrite(reader *bufio.Reader) {
	fmt.Println("Print filename: ")
	filename, err := reader.ReadString('\n')
	filename = strings.TrimSpace(filename)
	
	created, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error during creating file %s: %v\n", filename, err)
	}
	defer created.Close()
	
	if filepath.Ext(filename) == ".txt" {
		fmt.Println("\nPrint text: ")
		text, err := reader.ReadString('\n')
		words := strings.Fields(text)
		
		for _, w := range words {
			_, err = created.WriteString(w + " ")
			if err != nil {
				fmt.Printf("Error during writing into file %s: %v\n", filename, err)
			}
		}
	}

}


func DeleteFile(reader *bufio.Reader) {
	fmt.Println("Print filename: ")
	filename, _ := reader.ReadString('\n')
	filename = strings.TrimSpace(filename)
	
	err := os.Remove(filename)

	if err != nil {
		fmt.Printf("Error during deleting file %s: %v", filename, err)
	} else {
		fmt.Printf("File %s was deleted successfully", filename)
	}
}

func FillStudent(student *Student, reader *bufio.Reader) {
	fmt.Println("Print name: ")
	name, _ := reader.ReadString('\n')
	student.Name = strings.TrimSpace(name)

	fmt.Println("\nPrint surname:")
	surname, _ := reader.ReadString('\n')
	student.Surname = strings.TrimSpace(surname)

	fmt.Println("\nPrint age: ")
	ageInput, _ := reader.ReadString('\n')
	age, err := strconv.Atoi(strings.TrimSpace(ageInput))

	if err != nil {
		fmt.Printf("Input error: %v", err)
		student.Age = 0
	} else {
		student.Age = age
	}

	fmt.Println("\nPrint group: ")
	group, _ := reader.ReadString('\n')
	student.Group = strings.TrimSpace(group)
}

func CreateJSON(student Student, reader *bufio.Reader) {
	fmt.Println("Print filename: ")
	filename, _ := reader.ReadString('\n')
	filename = strings.TrimSpace(filename)
	
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error during creating file %s: %v", filename, err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(student); err != nil {
		fmt.Printf("Json dump error: %v", err)
	}

}

func AddZip(zipPath string, files ...string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", f)
		}

		// if strings.ToLower(filepath.Ext(f)) != ".txt" {
		// 	return fmt.Errorf("invalid extension for file %s", f)
		// }

		fileToZip, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("open %s: %w", f, err)
		}

		fi, err := fileToZip.Stat()
		if err != nil {
			fileToZip.Close()
			return fmt.Errorf("stat %s: %w", f, err)
		}

		header, err := zip.FileInfoHeader(fi)
		if err != nil {
			fileToZip.Close()
			return fmt.Errorf("fileinfoheader %s: %w", f, err)
		}

		header.Name = filepath.Base(f)
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			fileToZip.Close()
			return fmt.Errorf("create header %s: %w", f, err)
		}

		if _, err = io.Copy(writer, fileToZip); err != nil {
			fileToZip.Close()
			return fmt.Errorf("copy %s: %w", f, err)
		}

		if err := fileToZip.Close(); err != nil {
			return fmt.Errorf("close %s: %w", f, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("close zip writer: %w", err)
	}

	return nil
}

func DeleteZip(filename string) {
	if strings.ToLower(filepath.Ext(filename)) != ".zip" {
		fmt.Printf("Error: invalid file extension for %s\n", filename)
		return
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("Error: file %s does not exist\n", filename)
		return
	}

	if err := os.Remove(filename); err != nil {
		fmt.Printf("Error: failed to delete file %s: %v\n", filename, err)
		return
	}

	fmt.Printf("File %s successfully deleted\n", filename)
}

func ExtractZip(zipPath, destDir string) {
	if strings.ToLower(filepath.Ext(zipPath)) != ".zip" {
		fmt.Printf("Error: invalid file extension for %s\n", zipPath)
		return
	}

	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		fmt.Printf("Error: file %s does not exist\n", zipPath)
		return
	}

	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		fmt.Printf("Error: failed to resolve destination path: %v\n", err)
		return
	}

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: failed to get working directory: %v\n", err)
		return
	}

	if !strings.HasPrefix(absDestDir, filepath.Clean(workDir)+string(os.PathSeparator)) {
		fmt.Printf("Error: destination directory %s is outside working directory\n", destDir)
		return
	}

	if err := os.MkdirAll(absDestDir, 0755); err != nil {
		fmt.Printf("Error: failed to create destination directory: %v\n", err)
		return
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		fmt.Printf("Error: failed to open zip file %s: %v\n", zipPath, err)
		return
	}
	defer r.Close()

	// проверка на zip-бомбу через кол-во файлов
	if len(r.File) > maxFilesInArchive {
		fmt.Printf("Error: zip bomb detected — too many files (%d)\n", len(r.File))
		return
	}

	var totalSize int64 = 0

	for _, f := range r.File {
		fPath := filepath.Join(absDestDir, f.Name)

		if !strings.HasPrefix(filepath.Clean(fPath), absDestDir+string(os.PathSeparator)) {
			fmt.Printf("Error: illegal file path detected in archive (zip slip): %s\n", f.Name)
			return
		}

		// проверка на zip-бомбу через размер файла
		totalSize += int64(f.UncompressedSize64)
		if totalSize > int64(maxUncompressedSize) {
			fmt.Printf("Error: zip bomb detected — total uncompressed size too large (%d bytes)\n", totalSize)
			return
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fPath, f.Mode()); err != nil {
				fmt.Printf("Error: failed to create directory %s: %v\n", fPath, err)
				return
			}
			// continue
		}

		if err := os.MkdirAll(filepath.Dir(fPath), 0755); err != nil {
			fmt.Printf("Error: failed to create directory for %s: %v\n", fPath, err)
			return
		}

		rc, err := f.Open()
		if err != nil {
			fmt.Printf("Error: failed to open file %s inside archive: %v\n", f.Name, err)
			return
		}

		outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			fmt.Printf("Error: failed to create file %s: %v\n", fPath, err)
			return
		}

		if _, err = io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			fmt.Printf("Error: failed to extract file %s: %v\n", f.Name, err)
			return
		}

		rc.Close()
		outFile.Close()
	}

	fmt.Printf("Archive %s successfully extracted to %s\n", zipPath, absDestDir)
}

func CreateXML(reader *bufio.Reader) {
	fmt.Println("Print name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	
	fmt.Println("\nPrint age: ")
	age, _ := reader.ReadString('\n')
	age = strings.TrimSpace(age)
	
	ageI, err := strconv.Atoi(age)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	p := Person{Name: name, Age: ageI}
	xmlData, err := xml.MarshalIndent(p, "", " ")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	fmt.Println("\nPrint xml file name: ")
	filename, _ := reader.ReadString('\n')
	filename = strings.TrimSpace(filename)

	file, _ := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	defer file.Close()

	_, err = file.Write(xmlData)
	if err != nil {
		fmt.Printf("error writing to file: %v\n", err)
		return
	}

	// encoder := xml.NewEncoder(file)
	// encoder.Encode(xmlData)
}


