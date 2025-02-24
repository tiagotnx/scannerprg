# PRGScanner

PRGScanner é uma ferramenta de linha de comando (CLI) escrita em Go para analisar arquivos `.prg` e identificar funções e procedures não utilizadas. Ele percorre recursivamente um diretório, extrai as declarações de funções, verifica onde são utilizadas e gera um log detalhado com as funções/procedures não referenciadas.

## 📌 Funcionalidades

- 🔍 **Análise Recursiva**: Percorre diretórios buscando arquivos `.prg`.
- 📄 **Identificação de Declarações**: Detecta funções e procedures, diferenciando estáticas das globais.
- 📊 **Verificação de Uso**: Conta quantas vezes cada função/procedure é utilizada.
- 📜 **Geração de Log**: Cria um relatório detalhado, agrupando as funções/procedures não utilizadas por diretório e arquivo.
- ⏳ **Otimização Concorrente**: Usa múltiplas goroutines para melhorar o desempenho.
- 📈 **Estatísticas**: Apresenta porcentagens de uso e tempo total de processamento.
- ⚙️ **Configuração Flexível**: Pode ser configurado via variáveis de ambiente.

---

## 🚀 Instalação

1. **Clone o repositório**:
   ```bash
   git clone https://github.com/seu-usuario/PRGScanner.git
   cd PRGScanner
   ```

2. **Instale as dependências**:
   ```bash
   go mod tidy
   ```

3. **Compile o projeto**:
   ```bash
   go build -o prgscanner ./cmd/prgscanner
   ```

---

## 🛠 Uso

Após compilar, execute o programa da seguinte forma:

```bash
./prgscanner -dir="/caminho/do/codigo" -out="resultado.log"
```

### 🔹 Opções da CLI:

| Parâmetro       | Descrição                                       | Padrão         |
|----------------|----------------------------------------------|---------------|
| `-dir`        | Caminho do diretório a ser analisado        | `.` (atual)  |
| `-out`        | Nome do arquivo de log gerado               | `unused.log` |

#### 🔹 Exemplo de Execução:
```bash
./prgscanner -dir="/projetos" -out="analise.log"
```

---

## ⚙️ configuração via variáveis de ambiente

O PRGScanner pode ser configurado usando variáveis de ambiente. Exemplo:

```
PRGSCANNER_DIR  "/projetos"
PRGSCANNER_OUT  "analise.log"
```

---

## 📊 Exemplo de Log Gerado

```plaintext
========================================
   Log de Funções/Procedures Não Utilizadas
   Data: 2025-02-24 15:30:00
========================================

Estatísticas:
Tempo Total de Processamento: 1.23s
Total de Funções Globais: 50, Utilizadas: 35 (70.00%), Não Utilizadas: 15
Total de Funções Estáticas: 30, Utilizadas: 20 (66.67%), Não Utilizadas: 10

Funções/Procedures Globais Não Utilizadas (15):
------------------------------------------------------------
Diretório: /projetos/modulo1
  Arquivo: main.prg
    - Func1
    - Func2
  Arquivo: utils.prg
    - FuncAux
```

---

## 📜 Licença

Este projeto é distribuído sob a licença MIT. Consulte o arquivo `LICENSE` para mais informações.

---

## 🤝 Contribuição

Contribuições são bem-vindas! Siga os passos:

1. **Fork o repositório**.
2. **Crie uma branch** para sua feature: `git checkout -b feature/nova-feature`
3. **Commit suas alterações**: `git commit -m "Adiciona nova feature"`
4. **Faça o push para a branch**: `git push origin feature/nova-feature`
5. **Abra um Pull Request**.

---

## 📩 Contato

Dúvidas ou sugestões? Entre em contato:

📧 **Email**: tiagotnx@gmail.com  
🐙 **GitHub**: [tiagotnx](https://github.com/tiagotnx)  
📌 **LinkedIn**: [Tiago Nascimento da Silva](https://www.linkedin.com/in/tiagotnx/)  
```
