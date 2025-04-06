package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"
	_ "image/jpeg"

	"github.com/nfnt/resize"
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

// https://gerrit.googlesource.com/plugins/avatars-gravatar/+/b687eb0b55d464fea200b88059db1c393a1ad1ae/src/main/java/com/googlesource/gerrit/plugins/avatars/gravatar/GravatarAvatarProvider.java#101
func gravatarMD5JPG(email string) string {
	hasher := md5.Sum([]byte(strings.TrimSpace(email)))
	return hex.EncodeToString(hasher[:]) + ".jpg"
}

var gravatarHashes = []func(string) string{
	gravatarSHA256,
	gravatarMD5,
	gravatarMD5JPG,
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

		targets := make([]string, 0, len(suffixes)*len(gravatarHashes)+1)
		if file.Name() == "404.html" {
			targets = append(targets, file.Name())
		} else {
			name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			targets = append(targets, name+".png")

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

		img, format, err := image.Decode(bytes.NewReader(avatar))
		if err != nil {
			log.Printf("Decoding %q: %v", file.Name(), err)
		} else {
			log.Printf("%q decoded as %q", file.Name(), format)
			img = resize.Thumbnail(256, 256, img, resize.Lanczos3)

			var b bytes.Buffer
			err := png.Encode(&b, img)
			if err != nil {
				log.Printf("Encoding %q: %v", file.Name(), err)
			} else {
				avatar = b.Bytes()
			}
		}

		for _, target := range targets {
			err := os.WriteFile(filepath.Join(gravatarDir, target), avatar, 0o644)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%q transformed to %q", file.Name(), target)
		}
	}
}
