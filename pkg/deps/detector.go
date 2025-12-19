// pkg/deps/detector.go
package deps

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ModuleInfo struct {
	Name         string
	Version      string
	HasXS        bool
	XSFiles      []string
	PurePerl     []string
	Dependencies []string
	LocalPath    string
}

// AnalyzeModule анализирует модуль и возвращает информацию о нём
func AnalyzeModule(moduleName string) (*ModuleInfo, error) {
	// Ищем локально
	if info := findLocal(moduleName); info != nil {
		return info, nil
	}

	// Скачиваем с CPAN
	path, err := downloadFromCPAN(moduleName)
	if err != nil {
		return nil, err
	}

	// Анализируем содержимое
	info := &ModuleInfo{Name: moduleName, LocalPath: path}

	err = filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if f.IsDir() {
			return nil
		}

		switch filepath.Ext(p) {
		case ".xs":
			info.HasXS = true
			info.XSFiles = append(info.XSFiles, p)
		case ".pm":
			info.PurePerl = append(info.PurePerl, p)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Парсим зависимости
	info.Dependencies = parseDependencies(path)

	return info, nil
}

// findLocal ищет модуль в локальных путях
func findLocal(moduleName string) *ModuleInfo {
	// Конвертируем Module::Name в путь Module/Name
	relPath := strings.ReplaceAll(moduleName, "::", "/")

	searchPaths := []string{
		".",
		"lib",
		"local/lib/perl5",
	}

	// Добавляем пути из PERL5LIB
	if perl5lib := os.Getenv("PERL5LIB"); perl5lib != "" {
		searchPaths = append(searchPaths, strings.Split(perl5lib, ":")...)
	}

	for _, base := range searchPaths {
		pmPath := filepath.Join(base, relPath+".pm")
		if _, err := os.Stat(pmPath); err == nil {
			info := &ModuleInfo{
				Name:      moduleName,
				LocalPath: base,
			}
			info.PurePerl = append(info.PurePerl, pmPath)

			// Проверяем наличие XS
			xsPath := filepath.Join(base, relPath+".xs")
			if _, err := os.Stat(xsPath); err == nil {
				info.HasXS = true
				info.XSFiles = append(info.XSFiles, xsPath)
			}

			return info
		}
	}

	return nil
}

// downloadFromCPAN скачивает модуль с CPAN
func downloadFromCPAN(moduleName string) (string, error) {
	// Получаем URL для скачивания через MetaCPAN API
	apiURL := fmt.Sprintf("https://fastapi.metacpan.org/v1/download_url/%s", moduleName)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to query MetaCPAN: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("module %s not found on CPAN", moduleName)
	}

	var result struct {
		DownloadURL string `json:"download_url"`
		Version     string `json:"version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse MetaCPAN response: %w", err)
	}

	// Создаём директорию для кэша
	cacheDir := getCacheDir()
	moduleDir := filepath.Join(cacheDir, "modules", moduleName)
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return "", err
	}

	// Скачиваем архив
	tarPath := filepath.Join(moduleDir, "module.tar.gz")
	if err := downloadFile(result.DownloadURL, tarPath); err != nil {
		return "", err
	}

	// Распаковываем
	extractDir := filepath.Join(moduleDir, "src")
	if err := extractTarGz(tarPath, extractDir); err != nil {
		return "", err
	}

	return extractDir, nil
}

// downloadFile скачивает файл по URL
func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractTarGz распаковывает tar.gz архив
func extractTarGz(tarPath, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Открываем файл
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Распаковываем gzip
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	// Читаем tar
	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Целевой путь
		target := filepath.Join(destDir, header.Name)

		// Защита от path traversal
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Создаём директорию если нужно
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			// Создаём файл
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}

			// Копируем содержимое
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

// parseDependencies парсит зависимости из META.json или Makefile.PL
func parseDependencies(modulePath string) []string {
	var deps []string

	// Пробуем META.json
	metaPath := filepath.Join(modulePath, "META.json")
	if data, err := os.ReadFile(metaPath); err == nil {
		deps = append(deps, parseMetaJSON(data)...)
		return deps
	}

	// Пробуем META.yml
	metaYmlPath := filepath.Join(modulePath, "META.yml")
	if data, err := os.ReadFile(metaYmlPath); err == nil {
		deps = append(deps, parseMetaYAML(data)...)
		return deps
	}

	// Пробуем MYMETA.json
	mymetaPath := filepath.Join(modulePath, "MYMETA.json")
	if data, err := os.ReadFile(mymetaPath); err == nil {
		deps = append(deps, parseMetaJSON(data)...)
		return deps
	}

	return deps
}

// parseMetaJSON парсит зависимости из META.json
func parseMetaJSON(data []byte) []string {
	var meta struct {
		Prereqs struct {
			Runtime struct {
				Requires map[string]string `json:"requires"`
			} `json:"runtime"`
			Build struct {
				Requires map[string]string `json:"requires"`
			} `json:"build"`
			Configure struct {
				Requires map[string]string `json:"requires"`
			} `json:"configure"`
		} `json:"prereqs"`
	}

	var deps []string

	if err := json.Unmarshal(data, &meta); err != nil {
		return deps
	}

	for dep := range meta.Prereqs.Runtime.Requires {
		if !isCoreMod(dep) {
			deps = append(deps, dep)
		}
	}

	return deps
}

// parseMetaYAML парсит зависимости из META.yml (упрощённо)
func parseMetaYAML(data []byte) []string {
	var deps []string

	lines := strings.Split(string(data), "\n")
	inRequires := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "requires:") {
			inRequires = true
			continue
		}

		if inRequires {
			if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				inRequires = false
				continue
			}

			// Формат: "  Module::Name: version"
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) >= 1 {
				dep := strings.TrimSpace(parts[0])
				if dep != "" && !isCoreMod(dep) {
					deps = append(deps, dep)
				}
			}
		}
	}

	return deps
}

// isCoreMod проверяет, является ли модуль частью ядра Perl
func isCoreMod(name string) bool {
	coreMods := map[string]bool{
		"perl":                true,
		"strict":              true,
		"warnings":            true,
		"Exporter":            true,
		"Carp":                true,
		"File::Spec":          true,
		"File::Path":          true,
		"Data::Dumper":        true,
		"Storable":            true,
		"Getopt::Long":        true,
		"Pod::Usage":          true,
		"Test::More":          true,
		"Test::Simple":        true,
		"ExtUtils::MakeMaker": true,
	}
	return coreMods[name]
}

// getCacheDir возвращает директорию для кэша
func getCacheDir() string {
	// Сначала проверяем переменную окружения
	if cacheDir := os.Getenv("PERLC_CACHE"); cacheDir != "" {
		return cacheDir
	}

	// Используем домашнюю директорию
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	return filepath.Join(home, ".perlc", "cache")
}

// GetCachedModule возвращает закэшированный Go код модуля
func GetCachedModule(moduleName, version string) (string, bool) {
	cacheDir := getCacheDir()
	cachedFile := filepath.Join(cacheDir, "compiled", moduleName, version+".go")

	data, err := os.ReadFile(cachedFile)
	if err != nil {
		return "", false
	}

	return string(data), true
}

// CacheModule сохраняет скомпилированный модуль в кэш
func CacheModule(moduleName, version, goCode string) error {
	cacheDir := getCacheDir()
	dir := filepath.Join(cacheDir, "compiled", moduleName)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	cachedFile := filepath.Join(dir, version+".go")
	return os.WriteFile(cachedFile, []byte(goCode), 0644)
}
