package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	avatarDir   = "avatar"
	gravatarDir = "gravatar"
)

var suffixes = []string{
	"@circlecat.org",
	"@u.circlecat.org",
}

// https://docs.gravatar.com/api/avatars/go/
func gravatarSHA256(email string) string {
	hasher := sha256.Sum256([]byte(strings.TrimSpace(email)))
	return hex.EncodeToString(hasher[:])
}

// https://github.com/Automattic/go-gravatar/blob/master/gravatar.go
func gravatarMD5(email string) string {
	hasher := md5.Sum([]byte(strings.TrimSpace(email)))
	return hex.EncodeToString(hasher[:])
}

var gravatarHashes = []func(string) string{
	gravatarSHA256,
	gravatarMD5,
}

func main() {
	err := os.MkdirAll(gravatarDir, 0o755)
	if err != nil {
		log.Fatal(err)
	}

	files, err := os.ReadDir(avatarDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			log.Printf("Skipping directory %q", file.Name())
			continue
		}

		targets := make([]string, 0, len(suffixes)*len(gravatarHashes))
		if file.Name() == "404.html" {
			targets = append(targets, file.Name())
		} else {
			name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

			for _, suffix := range suffixes {
				email := name + suffix
				for _, gravatarHash := range gravatarHashes {
					hash := gravatarHash(email)
					log.Printf("%q translated to %q", email, hash)
					targets = append(targets, hash)
				}
			}
		}

		avatar, err := os.ReadFile(filepath.Join(avatarDir, file.Name()))
		if err != nil {
			log.Fatal(err)
		}

		for _, target := range targets {
			err := os.WriteFile(filepath.Join(gravatarDir, target), avatar, 0o644)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%q copied to %q", file.Name(), target)
		}
	}
}
