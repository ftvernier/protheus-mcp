# Segurança

O Protheus MCP foi projetado como uma ponte de diagnóstico **read-only** entre clientes de IA compatíveis com MCP e a infraestrutura do ERP.

## Garantias de segurança da v0.1

O projeto não disponibiliza intencionalmente tools para:

- executar SQL arbitrário;
- modificar tabelas do ERP;
- executar `KILL` em sessões do banco;
- encerrar processos do sistema operacional;
- reiniciar serviços do Windows;
- editar arquivos de configuração;
- retornar senhas configuradas para o banco de dados.

As consultas SQL existentes são fixas no código e destinadas exclusivamente à coleta de informações operacionais e DMVs necessárias ao diagnóstico.

## Credenciais do banco

Utilize um login dedicado ao monitoramento e conceda somente as permissões mínimas necessárias para consultar as DMVs utilizadas pelo projeto.

Não reutilize credenciais privilegiadas da aplicação, contas administrativas ou `sa`.

Credenciais devem ser fornecidas por variáveis de ambiente e nunca adicionadas ao repositório.

## Dados enviados ao cliente de IA

Algumas tools podem retornar informações como hostname, nome do programa cliente, identificador da sessão e trechos de SQL em execução. O SQL retornado é limitado em tamanho, mas ainda pode conter informações do ambiente.

Antes de utilizar o projeto em produção, revise as políticas de privacidade e tratamento de dados do cliente MCP e do modelo de IA utilizado.

## Reportando uma vulnerabilidade

Não publique credenciais, logs de produção, endereços de banco, hostnames internos ou informações confidenciais do ERP em Issues públicas.

Caso encontre uma vulnerabilidade, prefira reportá-la de forma privada ao mantenedor do projeto.
