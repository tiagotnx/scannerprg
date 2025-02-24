package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/cheggaaa/pb/v3"
)

// FunctionDeclaration representa uma função ou procedure encontrada.
type FunctionDeclaration struct {
	Name       string
	File       string
	Static     bool
	UsageCount int64 // Inicia em 1 para contabilizar a declaração
}

// DeclarationInfo representa os dados para log (nome e arquivo).
type DeclarationInfo struct {
	Name string
	File string
}

// Statistics armazena estatísticas do processamento.
type Statistics struct {
	TotalGlobal           int
	UnusedGlobal          int
	TotalStatic           int
	UnusedStatic          int
	TotalProcessingTime   time.Duration
	GlobalUsagePercentage float64
	StaticUsagePercentage float64
}

var (
	globalFunctions = make(map[string]*FunctionDeclaration)
	staticFunctions = make(map[string]map[string]*FunctionDeclaration)
	declMutex       sync.Mutex // Protege os mapas durante as declarações
)

// Regex para identificar declarações de função/procedure.
var declRegex = regexp.MustCompile(`(?i)^\s*(static\s+)?(function|procedure)\s+([a-zA-Z0-9_]+)`)

// SearchPRGFiles percorre recursivamente o diretório e retorna os caminhos dos arquivos .prg.
func SearchPRGFiles(root string) ([]string, error) {
	var prgFiles []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".prg") {
			prgFiles = append(prgFiles, path)
		}
		return nil
	})
	return prgFiles, err
}

// ProcessDeclarationsConcurrently processa as declarações de forma concorrente, atualizando a barra de progresso.
func ProcessDeclarationsConcurrently(files []string, bar *pb.ProgressBar) {
	numWorkers := runtime.NumCPU()
	fileCh := make(chan string, len(files))
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for file := range fileCh {
			processFileForDeclarations(file)
			bar.Increment()
		}
	}

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker()
	}
	for _, file := range files {
		fileCh <- file
	}
	close(fileCh)
	wg.Wait()
}

func processFileForDeclarations(file string) {
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Erro ao ler o arquivo %s: %v\n", file, err)
		return
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		match := declRegex.FindStringSubmatch(line)
		if match != nil {
			isStatic := strings.TrimSpace(match[1]) != ""
			name := match[3]
			declMutex.Lock()
			if isStatic {
				if staticFunctions[file] == nil {
					staticFunctions[file] = make(map[string]*FunctionDeclaration)
				}
				staticFunctions[file][name] = &FunctionDeclaration{
					Name:       name,
					File:       file,
					Static:     true,
					UsageCount: 1,
				}
			} else {
				if gf, exists := globalFunctions[name]; exists {
					gf.UsageCount++ // Em caso de declarações duplicadas
				} else {
					globalFunctions[name] = &FunctionDeclaration{
						Name:       name,
						File:       file,
						Static:     false,
						UsageCount: 1,
					}
				}
			}
			declMutex.Unlock()
		}
	}
}

// ProcessUsageConcurrently processa a verificação de uso de forma concorrente, atualizando a barra de progresso.
func ProcessUsageConcurrently(files []string, bar *pb.ProgressBar) {
	numWorkers := runtime.NumCPU()
	fileCh := make(chan string, len(files))
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for file := range fileCh {
			processFileForUsage(file)
			bar.Increment()
		}
	}

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker()
	}
	for _, file := range files {
		fileCh <- file
	}
	close(fileCh)
	wg.Wait()
}

func processFileForUsage(file string) {
	contentBytes, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Erro ao ler o arquivo %s: %v\n", file, err)
		return
	}
	content := strings.ToLower(string(contentBytes))
	freq := make(map[string]int)

	// Tokenização: letras, dígitos e underscore.
	tokens := strings.FieldsFunc(content, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_')
	})
	for _, token := range tokens {
		freq[token]++
	}

	// Atualiza o uso para funções globais
	for _, decl := range globalFunctions {
		lowerName := strings.ToLower(decl.Name)
		count := freq[lowerName]
		if file == decl.File && count > 0 {
			count-- // Desconta a declaração
		}
		if count > 0 {
			atomic.AddInt64(&decl.UsageCount, int64(count))
		}
	}

	// Atualiza o uso para funções estáticas
	if funcs, ok := staticFunctions[file]; ok {
		for _, decl := range funcs {
			lowerName := strings.ToLower(decl.Name)
			count := freq[lowerName]
			if count > 0 {
				count-- // Desconta a declaração
			}
			if count > 0 {
				atomic.AddInt64(&decl.UsageCount, int64(count))
			}
		}
	}
}

// GetUnusedDeclarations retorna slices com as declarações não utilizadas.
func GetUnusedDeclarations() (unusedGlobal []DeclarationInfo, unusedStatic []DeclarationInfo) {
	for _, decl := range globalFunctions {
		if atomic.LoadInt64(&decl.UsageCount) <= 1 {
			unusedGlobal = append(unusedGlobal, DeclarationInfo{
				Name: decl.Name,
				File: decl.File,
			})
		}
	}
	for file, funcs := range staticFunctions {
		for _, decl := range funcs {
			if atomic.LoadInt64(&decl.UsageCount) <= 1 {
				unusedStatic = append(unusedStatic, DeclarationInfo{
					Name: decl.Name,
					File: file,
				})
			}
		}
	}
	return
}

// CalculateStatistics calcula estatísticas com base nas declarações e no tempo de processamento.
func CalculateStatistics(totalProcessingTime time.Duration) Statistics {
	totalGlobal := len(globalFunctions)
	unusedGlobalCount := 0
	for _, decl := range globalFunctions {
		if decl.UsageCount <= 1 {
			unusedGlobalCount++
		}
	}

	totalStatic := 0
	unusedStaticCount := 0
	for _, funcs := range staticFunctions {
		for _, decl := range funcs {
			totalStatic++
			if decl.UsageCount <= 1 {
				unusedStaticCount++
			}
		}
	}

	globalUsed := totalGlobal - unusedGlobalCount
	staticUsed := totalStatic - unusedStaticCount

	var globalUsagePercentage, staticUsagePercentage float64
	if totalGlobal > 0 {
		globalUsagePercentage = float64(globalUsed) / float64(totalGlobal) * 100.0
	}
	if totalStatic > 0 {
		staticUsagePercentage = float64(staticUsed) / float64(totalStatic) * 100.0
	}

	return Statistics{
		TotalGlobal:           totalGlobal,
		UnusedGlobal:          unusedGlobalCount,
		TotalStatic:           totalStatic,
		UnusedStatic:          unusedStaticCount,
		TotalProcessingTime:   totalProcessingTime,
		GlobalUsagePercentage: globalUsagePercentage,
		StaticUsagePercentage: staticUsagePercentage,
	}
}

// GenerateLog gera um log organizado com agrupamento por diretório/arquivo e inclui estatísticas.
func GenerateLog(outputPath string, unusedGlobal []DeclarationInfo, unusedStatic []DeclarationInfo, stats Statistics) error {
	logFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	// Cabeçalho
	now := time.Now().Format("2006-01-02 15:04:05")
	header := fmt.Sprintf("========================================\n"+
		"   Log de Funções/Procedures Não Utilizadas\n"+
		"   Data: %s\n"+
		"========================================\n\n", now)
	logFile.WriteString(header)

	// Resumo de estatísticas
	summary := fmt.Sprintf("Estatísticas:\n"+
		"Tempo Total de Processamento: %s\n"+
		"Total de Funções Globais: %d, Utilizadas: %d (%.2f%%), Não Utilizadas: %d\n"+
		"Total de Funções Estáticas: %d, Utilizadas: %d (%.2f%%), Não Utilizadas: %d\n\n",
		stats.TotalProcessingTime,
		stats.TotalGlobal, stats.TotalGlobal-stats.UnusedGlobal, stats.GlobalUsagePercentage, stats.UnusedGlobal,
		stats.TotalStatic, stats.TotalStatic-stats.UnusedStatic, stats.StaticUsagePercentage, stats.UnusedStatic)
	logFile.WriteString(summary)

	// Agrupamento de declarações globais por diretório e arquivo
	globalGroup := make(map[string]map[string][]string)
	for _, decl := range unusedGlobal {
		dir := filepath.Dir(decl.File)
		fileBase := filepath.Base(decl.File)
		if globalGroup[dir] == nil {
			globalGroup[dir] = make(map[string][]string)
		}
		globalGroup[dir][fileBase] = append(globalGroup[dir][fileBase], decl.Name)
	}

	logFile.WriteString(fmt.Sprintf("Funções/Procedures Globais Não Utilizadas (%d):\n", len(unusedGlobal)))
	logFile.WriteString("============================================================\n")
	for dir, files := range globalGroup {
		logFile.WriteString("------------------------------------------------------------\n")
		logFile.WriteString(fmt.Sprintf("Diretório: %s\n", dir))
		logFile.WriteString("------------------------------------------------------------\n")
		for file, names := range files {
			logFile.WriteString(fmt.Sprintf("  Arquivo: %s\n", file))
			for _, name := range names {
				logFile.WriteString(fmt.Sprintf("    - %s\n", name))
			}
		}
		logFile.WriteString("\n")
	}
	logFile.WriteString("\n")

	// Agrupamento de declarações estáticas por diretório e arquivo
	staticGroup := make(map[string]map[string][]string)
	for _, decl := range unusedStatic {
		dir := filepath.Dir(decl.File)
		fileBase := filepath.Base(decl.File)
		if staticGroup[dir] == nil {
			staticGroup[dir] = make(map[string][]string)
		}
		staticGroup[dir][fileBase] = append(staticGroup[dir][fileBase], decl.Name)
	}

	logFile.WriteString(fmt.Sprintf("Funções/Procedures Estáticas Não Utilizadas (%d):\n", len(unusedStatic)))
	logFile.WriteString("============================================================\n")
	for dir, files := range staticGroup {
		logFile.WriteString("------------------------------------------------------------\n")
		logFile.WriteString(fmt.Sprintf("Diretório: %s\n", dir))
		logFile.WriteString("------------------------------------------------------------\n")
		for file, names := range files {
			logFile.WriteString(fmt.Sprintf("  Arquivo: %s\n", file))
			for _, name := range names {
				logFile.WriteString(fmt.Sprintf("    - %s\n", name))
			}
		}
		logFile.WriteString("\n")
	}

	return nil
}
