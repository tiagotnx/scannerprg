package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/tiagotnx/scannerprg/internal/scanner"
)

func main() {
	// Configuração via variáveis de ambiente
	dirDefault := "."
	if envDir := os.Getenv("PRGSCANNER_DIR"); envDir != "" {
		dirDefault = envDir
	}
	outDefault := "unused.log"
	if envOut := os.Getenv("PRGSCANNER_OUT"); envOut != "" {
		outDefault = envOut
	}

	// Flags com valores padrão (possivelmente sobrescritos pelas variáveis de ambiente)
	dirPtr := flag.String("dir", dirDefault, "diretório (ou unidade) a ser percorrido")
	outputPtr := flag.String("out", outDefault, "arquivo de log de saída")
	flag.Parse()

	startTime := time.Now()

	// Etapa 1: Busca dos arquivos .prg
	fmt.Println("Etapa 1: Buscando arquivos .prg...")
	prgFiles, err := scanner.SearchPRGFiles(*dirPtr)
	if err != nil {
		log.Fatalf("Erro ao buscar arquivos: %v", err)
	}

	// Etapa 2: Extração concorrente das declarações
	fmt.Println("Etapa 2: Identificando declarações de funções/procedures...")
	barDecl := pb.StartNew(len(prgFiles))
	scanner.ProcessDeclarationsConcurrently(prgFiles, barDecl)
	barDecl.Finish()

	// Etapa 3: Verificação concorrente do uso das funções/procedures
	fmt.Println("Etapa 3: Verificando o uso das funções/procedures...")
	barUsage := pb.StartNew(len(prgFiles))
	scanner.ProcessUsageConcurrently(prgFiles, barUsage)
	barUsage.Finish()

	// Calcula as estatísticas do processamento
	totalTime := time.Since(startTime)
	stats := scanner.CalculateStatistics(totalTime)

	// Obtém as declarações não utilizadas
	unusedGlobal, unusedStatic := scanner.GetUnusedDeclarations()

	// Etapa 4: Geração do log com estatísticas e agrupamentos
	fmt.Println("Etapa 4: Gerando arquivo de log...")
	err = scanner.GenerateLog(*outputPtr, unusedGlobal, unusedStatic, stats)
	if err != nil {
		log.Fatalf("Erro ao gerar arquivo de log: %v", err)
	}

	fmt.Printf("Processamento concluído. Log gerado em: %s\n", *outputPtr)
}
