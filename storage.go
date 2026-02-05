package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

type Storage struct {
	dataDir string
}

func NewStorage(dataDir string) *Storage {
	return &Storage{dataDir: dataDir}
}

var usernameRegex = regexp.MustCompile(`^[a-z0-9]{3,20}$`)

func (s *Storage) ValidateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username must be 3-20 lowercase letters or numbers")
	}
	return nil
}

func (s *Storage) GenerateID() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *Storage) Save(username string, content io.Reader) (string, error) {
	if err := s.ValidateUsername(username); err != nil {
		return "", err
	}

	id, err := s.GenerateID()
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s-%s.html", username, id)
	filepath := filepath.Join(s.dataDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, content); err != nil {
		os.Remove(filepath)
		return "", err
	}

	return fmt.Sprintf("%s-%s", username, id), nil
}

func (s *Storage) Get(slug string) ([]byte, error) {
	filepath := filepath.Join(s.dataDir, slug+".html")
	return os.ReadFile(filepath)
}
