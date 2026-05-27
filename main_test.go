package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSafesearch_s(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected string
	}{
		{
			name:     "true input",
			input:    true,
			expected: "",
		},
		{
			name:     "false input",
			input:    false,
			expected: "off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safesearch_s(tt.input)
			if result != tt.expected {
				t.Errorf("safesearch_s(%v) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMoveFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-movefile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	// Create source file
	err = ioutil.WriteFile(srcPath, []byte("hello world"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test successful move
	err = moveFile(dstPath, srcPath)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify destination exists and has correct content
	content, err := ioutil.ReadFile(dstPath)
	if err != nil {
		t.Errorf("failed to read destination file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("destination content = %v; want %v", string(content), "hello world")
	}

	// Verify source is removed
	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Errorf("expected source file to be removed, got err: %v", err)
	}

	// Test moving non-existent source
	err = moveFile(filepath.Join(tmpDir, "non_existent_dest.txt"), filepath.Join(tmpDir, "non_existent_src.txt"))
	if err == nil {
		t.Error("expected error when moving non-existent file, got nil")
	}
}

func TestWorker(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/image.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("fake image data"))
		case "/image.gif":
			w.Header().Set("Content-Type", "image/gif")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("fake gif data"))
		case "/image.png":
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("fake png data"))
		case "/not_an_image":
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("this is not an image"))
		case "/404":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	tmpDir, err := ioutil.TempDir("", "test-worker-tmp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	outDir, err := ioutil.TempDir("", "test-worker-out")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outDir)

	q := make(chan string, 10)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var count int64 = 4 // Allow downloading 4 files

	wg.Add(1)
	go worker(&mu, &wg, q, tmpDir, outDir, &count)

	// Send URLs to the queue
	q <- server.URL + "/image.jpg"
	q <- server.URL + "/image.gif"
	q <- server.URL + "/image.png"
	q <- server.URL + "/not_an_image"
	q <- server.URL + "/404"
	q <- "http://invalid-url-that-does-not-exist.local"
	q <- server.URL + "/image.jpg" // A 4th image (allowed)
	q <- server.URL + "/image.jpg" // A 5th image (should not be saved, count limit, but downloaded)

	close(q)
	wg.Wait()

	// Verify downloads
	files, err := ioutil.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}

	// 4 successful image downloads where count wasn't < 0 before renaming
	if len(files) != 4 {
		t.Errorf("expected 4 files in outDir, found %d", len(files))
	}

	// We should check that a file for each type (jpg, gif, png) exists
	hasJpg := false
	hasGif := false
	hasPng := false

	for _, file := range files {
		content, err := ioutil.ReadFile(filepath.Join(outDir, file.Name()))
		if err != nil {
			t.Errorf("failed to read downloaded file: %v", err)
		}
		if strings.HasSuffix(file.Name(), ".jpg") {
			hasJpg = true
			if string(content) != "fake image data" {
				t.Errorf("downloaded file content = %v; want %v", string(content), "fake image data")
			}
		} else if strings.HasSuffix(file.Name(), ".gif") {
			hasGif = true
			if string(content) != "fake gif data" {
				t.Errorf("downloaded file content = %v; want %v", string(content), "fake gif data")
			}
		} else if strings.HasSuffix(file.Name(), ".png") {
			hasPng = true
			if string(content) != "fake png data" {
				t.Errorf("downloaded file content = %v; want %v", string(content), "fake png data")
			}
		} else {
			t.Errorf("unexpected file suffix: %s", file.Name())
		}
	}

	if !hasJpg || !hasGif || !hasPng {
		t.Errorf("expected at least one jpg, gif, and png file. jpg:%v gif:%v png:%v", hasJpg, hasGif, hasPng)
	}

	// Verify count was decremented correctly
	// Initially 4, decremented 5 times (5 valid images downloaded), so it should be -1
	if atomic.LoadInt64(&count) != -1 {
		t.Errorf("expected count to be -1, got %d", atomic.LoadInt64(&count))
	}
}

func TestMoveFile_Errors(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-movefile-err")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcPath := filepath.Join(tmpDir, "source.txt")
	err = ioutil.WriteFile(srcPath, []byte("hello"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test invalid destination
	err = moveFile(filepath.Join(tmpDir, "non_existent_dir", "dest.txt"), srcPath)
	if err == nil {
		t.Error("expected error when destination directory doesn't exist")
	}
}

func TestWorker_Errors(t *testing.T) {
	q := make(chan string, 1)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var count int64 = 1

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("fake image data"))
	}))
	defer server.Close()

	wg.Add(1)
	// Passing invalid tmpdir to force os.Create error
	go worker(&mu, &wg, q, "/path/that/does/not/exist", "/tmp", &count)

	q <- server.URL + "/image.jpg"
	close(q)
	wg.Wait()

	// Ensure count wasn't decremented because of the Create error
	if count != 1 {
		t.Errorf("expected count to be 1, got %d", count)
	}
}

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestMainWithArgs(t *testing.T) {
	outDir, err := ioutil.TempDir("", "test-main-out")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outDir)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-n", "1", "-o", outDir, "test", "query"}

	oldCommandLine := flag.CommandLine
	defer func() { flag.CommandLine = oldCommandLine }()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Mock HTTP transport
	originalTransport := http.DefaultClient.Transport
	defer func() { http.DefaultClient.Transport = originalTransport }()

	http.DefaultClient.Transport = &mockTransport{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.String(), "bing.com") {
				mockResponseBody := `murl&quot;:&quot;http://example.com/image.jpg&quot;`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(mockResponseBody)),
					Header:     make(http.Header),
				}, nil
			}
			header := make(http.Header)
			header.Set("Content-Type", "image/jpeg")
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(strings.NewReader("fake image")),
				Header:     header,
			}, nil
		},
	}

	main()
}
