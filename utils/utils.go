package utils

import(
	"strings"
	"runtime"
	"path/filepath"
	"os"
	"errors"
	"regexp"
	"bufio"
	"strconv"
)

func wasRunFromSrc() bool {
	buildPath := filepath.Join(os.TempDir(), "go-build")
	return strings.HasPrefix(os.Args[0], buildPath)
}

func getScriptDir() (string, error) {
	var (
		ok    bool
		err   error
		scriptDir string
	)
	runFromSrc := wasRunFromSrc()
	if runFromSrc {
		_, scriptDir, _, ok = runtime.Caller(0)
		if !ok {
			return "", errors.New("failed to get script directory")
		}
		scriptDir = scriptDir[:len(scriptDir)-14] // - "utils/utils.go"
	} else {
		scriptDir, err = os.Executable()
		if err != nil {
			return "", err
		}
	}
	return filepath.Dir(scriptDir), nil
}

func CDToScriptDir() error {
	scriptDir, err := getScriptDir()
	if err != nil {
		return err
	}
	return os.Chdir(scriptDir)
}

func Sanitise(fname string, isFolder bool) string {
	var regexStr string
	if isFolder {
		regexStr = `[:*?"><|]`
	} else {
		regexStr = `[\/:*?"><|]`
	}
	return regexp.MustCompile(regexStr).ReplaceAllString(fname, "_")
}

func MakeDirs(path string) error {
	return os.MkdirAll(path, 0777)
}

func readTxtFile(path string) ([]string, error) {
	var lines []string
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return lines, nil
}

func contains(lines []string, value string) bool {
	for _, line := range lines {
		if strings.EqualFold(line, value) {
			return true
		}
	}
	return false
}

func ProcessUrls(urls []string) ([]string, error) {
	var (
		processed []string
		txtPaths  []string
	)
	for _, _url := range urls {
		if strings.HasSuffix(_url, ".txt") && !contains(txtPaths, _url) {
			txtLines, err := readTxtFile(_url)
			if err != nil {
				return nil, err
			}
			for _, txtLine := range txtLines {
				if !contains(processed, txtLine) {
					processed = append(processed, txtLine)
				}
			}
			txtPaths = append(txtPaths, _url)
		} else {
			if !contains(processed, _url) {
				processed = append(processed, _url)
			}
		}
	}
	return processed, nil
}

func FileExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func FormatFreq(freq int) string {
	freqStr := strconv.Itoa(freq / 100)
	if strings.HasSuffix(freqStr, "0") {
		return strconv.Itoa(freq / 1000)
	} else {
		freqStrLen := len(freqStr)
		return freqStr[:freqStrLen-1] + "." + freqStr[freqStrLen-1:]
	}
}