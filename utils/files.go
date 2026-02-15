package utils

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// EnsureAssets guarantees that the required test asset files exist in the
// assets/ directory. If they are missing, it creates minimal valid files.
func EnsureAssets(assetsDir string) error {
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		return fmt.Errorf("create assets dir: %w", err)
	}

	imgPath := filepath.Join(assetsDir, "test_image.png")
	if err := createTestPNG(imgPath); err != nil {
		return fmt.Errorf("create test image: %w", err)
	}

	pdfPath := filepath.Join(assetsDir, "test_document.pdf")
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		if err := createTestPDF(pdfPath); err != nil {
			return fmt.Errorf("create test PDF: %w", err)
		}
	}

	return nil
}

// realEstateImageURLs is a curated list of royalty-free real-estate photos.
// A random one is picked on every run so the test image changes each time.
var realEstateImageURLs = []string{
	"https://images.unsplash.com/photo-1560518883-ce09059eeffa?w=400&q=80",
	"https://images.unsplash.com/photo-1570129477492-45c003edd2be?w=400&q=80",
	"https://images.unsplash.com/photo-1600596542815-ffad4c1539a9?w=400&q=80",
	"https://images.unsplash.com/photo-1600585154340-be6161a56a0c?w=400&q=80",
	"https://images.unsplash.com/photo-1512917774080-9991f1c4c750?w=400&q=80",
	"https://images.unsplash.com/photo-1580587771525-78b9dba3b914?w=400&q=80",
	"https://images.unsplash.com/photo-1564013799919-ab600027ffc6?w=400&q=80",
	"https://images.unsplash.com/photo-1568605114967-8130f3a36994?w=400&q=80",
	"https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?w=400&q=80",
	"https://images.unsplash.com/photo-1605276374104-dee2a0ed3cd6?w=400&q=80",
}

// createTestPNG downloads a random real-estate image from the internet
// and saves it to the given path.
func createTestPNG(path string) error {
	url := realEstateImageURLs[rand.Intn(len(realEstateImageURLs))]

	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download real-estate image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download real-estate image: unexpected status %s", resp.Status)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("write image to disk: %w", err)
	}

	return nil
}

// createTestPDF creates a minimal single-page PDF document.
func createTestPDF(path string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Test Document - Tarh Script")
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "This is an auto-generated test document for the workflow orchestrator.")
	return pdf.OutputFileAndClose(path)
}
