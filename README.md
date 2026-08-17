# Protheus MCP

> **Troubleshooting de ambientes TOTVS Protheus assistido por IA.**

**Protheus MCP** é um servidor open source baseado no **Model Context Protocol (MCP)** e desenvolvido em **Go**, criado para fornecer a assistentes de IA visibilidade **em tempo real e somente leitura** sobre ambientes TOTVS Protheus.

Em vez de perguntar para uma IA algo genérico como *“o que pode deixar o Protheus lento?”*, a proposta é permitir uma pergunta muito mais interessante:

> **“Meu Protheus está lento agora. Investigue o que está acontecendo.”**

A IA pode utilizar as ferramentas expostas pelo Protheus MCP para coletar evidências do Windows, dos processos do Protheus AppServer e do SQL Server e, a partir desses dados, auxiliar na investigação do problema.

![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)
![MCP](https://img.shields.io/badge/MCP-Server-5A45FF)
![Windows](https://img.shields.io/badge/Windows-principal-0078D4?logo=windows)
![SQL Server](https://img.shields.io/badge/SQL%20Server-suportado-CC2927?logo=microsoftsqlserver&logoColor=white)
![Read Only](https://img.shields.io/badge/segurança-read--only-success)
![License](https://img.shields.io/badge/licença-MIT-blue)

## Por que o Protheus MCP?

Investigar problemas de performance em um ERP normalmente exige alternar entre Gerenciador de Tarefas, serviços do Windows, DMVs do SQL Server, processos do AppServer, logs e ferramentas de monitoramento.

O Protheus MCP começa a criar uma ponte única entre esse contexto operacional e um assistente de IA compatível com MCP.

```text
                    Assistente de IA
                    compatível com MCP
                           │
                           │ MCP / stdio
                           ▼
                   ┌────────────────┐
                   │  Protheus MCP  │
                   │       Go       │
                   └───────┬────────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
           Windows      Protheus     SQL Server
                        AppServer
              │            │            │
              └────────────┼────────────┘
                           ▼
                    Dados operacionais
                           │
                           ▼
                  Diagnóstico assistido
                         por IA
```

## O que a v0.1.0-alpha faz?

A primeira versão pública disponibiliza **seis tools MCP read-only**:

| Tool | Finalidade |
| --- | --- |
| `get_system_health` | CPU, memória, uptime do sistema operacional e utilização dos discos locais |
| `get_protheus_processes` | Localiza processos do AppServer e retorna PID, CPU, memória e uptime |
| `get_database_health` | Valida conexão e latência do SQL Server e identifica requests ativos, bloqueados e de longa duração |
| `get_long_running_queries` | Lista requests atualmente em execução acima de um tempo mínimo configurável |
| `get_blocking_sessions` | Identifica relações de bloqueio entre sessões no SQL Server |
| `get_session_details` | Aprofunda a análise de uma sessão específica, incluindo waits, blocker, CPU e SQL em execução |

### Exemplo de investigação

```text
Usuário: Meu ambiente Protheus está lento. Investigue.

IA
 ├─ get_system_health()
 │    └─ CPU: normal / Memória: normal / Disco: normal
 │
 ├─ get_database_health()
 │    └─ blocked_requests: 4
 │
 ├─ get_blocking_sessions()
 │    └─ sessão 84 bloqueia as sessões 117, 132 e 141
 │
 └─ get_session_details(session_id=84)
      └─ TOTVS Application Server / request SQL de longa duração

IA: O principal indício neste momento é contenção no banco de dados.
    A sessão 84, originada pelo TOTVS Application Server, está bloqueando
    outras três sessões. Os recursos do host estão dentro da normalidade.
```

O MCP fornece as **evidências estruturadas**. O assistente de IA utiliza essas evidências para raciocinar, correlacionar sinais e explicar o diagnóstico.

## Segurança: read-only por design

A versão pública foi intencionalmente criada para diagnóstico e **não realiza ações corretivas automaticamente**.

Ela **não**:

- executa SQL arbitrário;
- executa `KILL` em sessões do SQL Server;
- encerra processos do AppServer;
- reinicia serviços do Windows;
- altera `appserver.ini`;
- altera dados do ERP ou do banco;
- retorna credenciais do banco através das tools MCP.

O texto SQL retornado pelas ferramentas de diagnóstico também é limitado para reduzir exposição desnecessária de dados e consumo de contexto.

> **Importante:** revise a política de privacidade e tratamento de dados do cliente de IA utilizado. Metadados operacionais e trechos de SQL retornados pelas tools podem fazer parte do contexto enviado ao modelo utilizado pelo cliente MCP.

## Requisitos

### Para compilar o projeto

- Go 1.23+
- Windows é o sistema operacional principal suportado nesta primeira alpha
- SQL Server é opcional: as tools de sistema e processos funcionam mesmo sem configuração do banco

### Para utilizar uma versão compilada

O Go não é necessário em runtime. Basta utilizar o executável Windows disponibilizado em uma release e configurá-lo no cliente MCP.

## Compilando no Windows

```powershell
go mod tidy
go test ./...
go build -o protheus-mcp.exe ./cmd/protheus-mcp
```

Ou utilize o script:

```powershell
.\scripts\build-windows.ps1
```

## Configuração

Nesta versão, a configuração é realizada por **variáveis de ambiente**, mantendo credenciais fora do código-fonte e do repositório.

| Variável | Obrigatória | Padrão | Descrição |
| --- | --- | --- | --- |
| `PROTHEUS_PROCESS` | Não | `appserver` | Nome/padrão utilizado para localizar processos do Protheus AppServer |
| `DB_HOST` | Para tools SQL | — | Host do SQL Server |
| `DB_PORT` | Não | `1433` | Porta TCP do SQL Server |
| `DB_NAME` | Para tools SQL | — | Nome do banco Protheus |
| `DB_USER` | Para tools SQL | — | Login utilizado para monitoramento |
| `DB_PASSWORD` | Para tools SQL | — | Senha do login de monitoramento |
| `DB_ENCRYPT` | Não | `true` | Habilita conexão criptografada com o SQL Server |
| `DB_TRUST_SERVER_CERTIFICATE` | Não | `false` | Aceita o certificado apresentado pelo SQL Server sem validar a cadeia da CA |
| `QUERY_TIMEOUT_SECONDS` | Não | `5` | Timeout das consultas de monitoramento |

### Exemplo

```cmd
set PROTHEUS_PROCESS=appserver
set DB_HOST=localhost
set DB_PORT=1433
set DB_NAME=PROTHEUS
set DB_USER=protheus_monitor
set DB_PASSWORD=ALTERE_AQUI
set DB_ENCRYPT=true
set DB_TRUST_SERVER_CERTIFICATE=false
```

Em ambientes SQL Server que utilizam certificado interno/self-signed, pode ser necessário:

```cmd
set DB_TRUST_SERVER_CERTIFICATE=true
```

Sempre que possível, prefira utilizar um certificado confiável para o cliente. Essa opção mantém a conexão criptografada, mas ignora a validação da cadeia do certificado.

### Privilégio mínimo

Utilize um login **dedicado exclusivamente ao monitoramento**.

Evite utilizar `sa`, a própria credencial utilizada pelo Protheus ou qualquer conta com privilégios de escrita desnecessários. A alpha executa somente consultas de diagnóstico e leitura das DMVs necessárias.

## Testando com o MCP Inspector

O **MCP Inspector** é uma ótima forma de validar o servidor antes de conectá-lo a um assistente de IA.

Exemplo utilizando o CMD do Windows:

```cmd
npx @modelcontextprotocol/inspector ^
  -e DB_HOST=localhost ^
  -e DB_PORT=1433 ^
  -e DB_NAME=PROTHEUS ^
  -e DB_USER=protheus_monitor ^
  -e DB_PASSWORD=ALTERE_AQUI ^
  -e DB_ENCRYPT=true ^
  -e DB_TRUST_SERVER_CERTIFICATE=false ^
  .\protheus-mcp.exe
```

Após conectar, as seis tools deverão aparecer no Inspector.

## Configurando em um cliente MCP

O Protheus MCP utiliza atualmente o transporte MCP via `stdio`.

Uma configuração típica de cliente MCP é semelhante a:

```json
{
  "mcpServers": {
    "protheus": {
      "command": "C:\\tools\\protheus-mcp\\protheus-mcp.exe",
      "env": {
        "PROTHEUS_PROCESS": "appserver",
        "DB_HOST": "localhost",
        "DB_PORT": "1433",
        "DB_NAME": "PROTHEUS",
        "DB_USER": "protheus_monitor",
        "DB_PASSWORD": "ALTERE_AQUI",
        "DB_ENCRYPT": "true",
        "DB_TRUST_SERVER_CERTIFICATE": "false"
      }
    }
  }
}
```

Como o servidor utiliza `stdio`, o `stdout` é reservado para o tráfego JSON-RPC do MCP. Logs da aplicação são enviados para `stderr`.

## Prompts para experimentar

Depois de conectar o servidor a um cliente de IA compatível com MCP, experimente:

```text
Analise a saúde atual do meu ambiente Protheus utilizando as ferramentas
read-only disponíveis. Explique as evidências encontradas antes de recomendar
o que devo investigar.
```

```text
Meu Protheus está lento agora. Investigue se o principal indício está no
servidor Windows, nos processos do AppServer ou no SQL Server.
```

```text
Verifique se existem bloqueios no SQL Server. Caso encontre um blocker,
analise a sessão responsável e explique o que ela está fazendo.
Não realize nenhuma ação corretiva.
```

## Estrutura do projeto

```text
protheus-mcp/
├── cmd/
│   └── protheus-mcp/
├── internal/
│   ├── config/
│   ├── database/
│   │   └── sqlserver/
│   ├── protheus/
│   ├── system/
│   └── tools/
├── scripts/
├── .github/workflows/
├── CONTRIBUTING.md
├── SECURITY.md
├── LICENSE
└── README.md
```

## Roadmap

### v0.1.x

- [x] MCP Server desenvolvido em Go
- [x] Health check do Windows
- [x] Descoberta de processos Protheus AppServer
- [x] Health check do SQL Server
- [x] Diagnóstico de queries de longa duração
- [x] Diagnóstico de sessões bloqueadas
- [x] Detalhamento de sessões SQL
- [x] Configuração de confiança do certificado TLS do SQL Server
- [ ] Ampliar cobertura de testes automatizados
- [ ] Melhorar experiência inicial de configuração

### v0.2

- [ ] Descoberta e status dos serviços Windows do Protheus
- [ ] Parser de `appserver.ini` e contexto dos environments
- [ ] Testes de conectividade com DBAccess, License Server e REST
- [ ] Diagnóstico de waits do SQL Server

### Futuro

- [ ] Provider PostgreSQL
- [ ] Diagnóstico de logs do AppServer
- [ ] Correlação entre múltiplos hosts
- [ ] Mais contexto operacional específico do Protheus

## Contribuindo

Issues, feedbacks, resultados de testes e Pull Requests são muito bem-vindos.

Como o projeto ainda está em alpha, testes em diferentes topologias Protheus serão especialmente importantes para definir as próximas funcionalidades.

Consulte também [`CONTRIBUTING.md`](CONTRIBUTING.md).

## Aviso

Este é um projeto open source independente e **não possui vínculo ou endosso da TOTVS S.A.** TOTVS e Protheus são marcas de seus respectivos proprietários.

O projeto ainda é experimental. Valide os diagnósticos em ambientes controlados antes de utilizá-los como base para decisões em produção.

## Licença

MIT — consulte [`LICENSE`](LICENSE).
