package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/blake2b"
)

type FileInfo struct {
	Size int64
	Hash string
}

var AllowedExtensions = []string{
	".mp3", ".wav", ".flac", ".ogg",
}

func calculateHash(fpath string) (string, error) {
	/* Calculate the hash of a file using the BLAKE2b algorithm.
	 * fpath: the path to the file to hash.
	 * Returns the hash as a string of hexadecimal digits.
	 * Returns an error if the file cannot be opened or read.
	 */
	hasher, err := blake2b.New256(nil)
	if err != nil {
		return "", err
	}

	file, err := os.Open(fpath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := make([]byte, 4096)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		hasher.Write(buf[:n])
	}

	hash := hasher.Sum(nil)
	return fmt.Sprintf("%x", hash), nil
}

func findDuplicates(directorioRaiz string) ([]string, error) {
	hashes := make(map[string]FileInfo)
	filesToCheck := []string{}
	dupes := []string{}

	err := filepath.Walk(directorioRaiz, func(ruta string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := filepath.Ext(ruta)
			isAllowed := false
			for _, allowedExt := range AllowedExtensions { // Iterate through allowed extensions
				if ext == allowedExt {
					isAllowed = true
					break // Exit loop once a match is found
				}
			}
			if isAllowed {
				filesToCheck = append(filesToCheck, ruta)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Now process filesToCheck
	for _, file := range filesToCheck {
		info, err := os.Stat(file)
		if err != nil {
			log.Printf("Error getting FileInfo: %v\n", err)
			continue // Skip to the next file
		}

		hash, err := calculateHash(file)
		if err != nil {
			log.Printf("Error calculating hash: %v\n", err)
			continue // Skip if hash calculation fails
		}

		if existingFileInfo, ok := hashes[file]; ok {
			if info.Size() == existingFileInfo.Size && hash == existingFileInfo.Hash {
				dupes = append(dupes, file)
			}
		} else {
			hashes[file] = FileInfo{Size: info.Size(), Hash: hash}
		}
	}

	return dupes, nil
}

func main() {
	// call findDuplicates with argument from command line
	if len(os.Args) != 2 {
		fmt.Println("Uso: gofindups <directorio>")
		os.Exit(1)
	}
	directorioRaiz := os.Args[1]
	dupes, err := findDuplicates(directorioRaiz)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(dupes) == 0 {
		fmt.Println("No se encontraron duplicados.")
		return
	}
	fmt.Println("Duplicados encontrados:")
	for _, dup := range dupes {
		fmt.Println(dup)
	}
	fmt.Println("Desea borrar los duplicados? (s/n)")
	var yesno rune
	_, err = fmt.Scanf("%c", &yesno)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if yesno == 's' || yesno == 'S' {
		for _, dup := range dupes {
			err := os.Remove(dup)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		fmt.Println("Duplicados eliminados.")
	} else {
		fmt.Println("Operación cancelada.")
	}
}
