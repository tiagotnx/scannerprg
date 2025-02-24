package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/cheggaaa/pb/v3"
	"github.com/tiagotnx/scannerprg/internal/scanner"
)

func main() {
	// Parâmetros da CLI
	dirPtr := flag.String("dir", ".", "diretório (ou unidade) a ser percorrido")
	outputPtr := flag.String("out", "unused.log", "arquivo de log de saída")
	flag.Parse()

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

	// Etapa 4: Geração do arquivo de log com os resultados
	unusedGlobal, unusedStatic := scanner.GetUnusedDeclarations()
	fmt.Println("Etapa 4: Gerando arquivo de log...")
	err = scanner.GenerateLog(*outputPtr, unusedGlobal, unusedStatic)
	if err != nil {
		log.Fatalf("Erro ao gerar arquivo de log: %v", err)
	}

	fmt.Printf("Processamento concluído. Log gerado em: %s\n", *outputPtr)
}
