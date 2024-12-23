package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/blake2b"
)

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
	hashes := make(map[string]string)
	dupes := []string{}
	err := filepath.Walk(directorioRaiz, func(ruta string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error al acceder a la ruta %q: %v\n", ruta, err)
			return nil
		}
		if !info.IsDir() && (filepath.Ext(ruta) == ".mp3" || filepath.Ext(ruta) == ".wav" || filepath.Ext(ruta) == ".flac") {
			hash, err := calculateHash(ruta)
			if err != nil {
				log.Printf("Error al calcular el hash del archivo %q: %v\n", ruta, err)
				return nil
			}

			if rutaExistente, ok := hashes[hash]; ok {
				// Verificar tamaño del archivo
				existingFileInfo, err := os.Stat(rutaExistente)
				if err != nil {
					return err
				}

				if info.Size() == existingFileInfo.Size() {
					if info.Size() == existingFileInfo.Size() {
						// Calcular hash completo si los tamaños coinciden
						hashCompleto1, _ := calculateHash(ruta)
						hashCompleto2, _ := calculateHash(rutaExistente)
						if hashCompleto1 == hashCompleto2 {
							fmt.Printf("Duplicado encontrado: %s y %s\n", ruta, rutaExistente)
							dupes = append(dupes, rutaExistente)
						}
					}
				}
			} else {
				hashes[hash] = ruta
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Error al recorrer el directorio %q: %v\n", directorioRaiz, err)
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
	} else {
		fmt.Println("Duplicados encontrados:")
		for _, dupe := range dupes {
			fmt.Println(dupe)
		}
	}
}
