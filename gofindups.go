package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/blake2b"
)

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
	filesBySize := make(map[int64][]string)
	dupes := []string{}
	hashes := make(map[string]string) // map from hash to file path

	err := filepath.Walk(directorioRaiz, func(ruta string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %q: %v\n", ruta, err)
		}

		if !info.IsDir() {
			ext := filepath.Ext(ruta)
			isAllowed := false
			for _, allowedExt := range AllowedExtensions {
				if ext == allowedExt {
					isAllowed = true
					break
				}
			}
			if isAllowed {
				filesBySize[info.Size()] = append(filesBySize[info.Size()], ruta)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}



	for _, files := range filesBySize {
		if len(files) < 2 {
			continue // no duplicates possible if less than 2 files of the same size
		}

		for _, file := range files {
			hash, err := calculateHash(file)
			if err != nil {
				log.Printf("Error calculating hash: %v\n", err)
				continue
			}

			if _, ok := hashes[hash]; ok {
				dupes = append(dupes, file)
			} else {
				hashes[hash] = file
			}
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
	fmt.Printf("Duplicados encontrados: %d\n", len(dupes))
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
				fmt.Printf("Error al borrar %s: %v\n", dup, err)
			} else {
				fmt.Printf("Borrado %s\n", dup)
			}
		}
	}
}
