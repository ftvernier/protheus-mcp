# Contribuindo

Contribuições são muito bem-vindas, principalmente diagnósticos reproduzíveis em ambientes TOTVS Protheus reais.

Ao adicionar novas tools, mantenha os seguintes princípios:

1. **read-only por padrão**;
2. escopo claro e específico;
3. saída estruturada para consumo por assistentes de IA;
4. segurança para uso em ambientes semelhantes a produção;
5. independência do provedor de LLM.

## Antes de abrir um Pull Request

Formate e valide o projeto:

```bash
gofmt -w ./cmd ./internal
go test ./...
```

Sempre que possível, descreva no Pull Request:

- qual problema a mudança resolve;
- como a funcionalidade foi testada;
- sistema operacional e banco utilizados no teste;
- exemplo de saída sem informações sensíveis.

## Segurança e privacidade

Nunca inclua em exemplos, testes, Issues ou Pull Requests:

- dados de clientes;
- usuários ou senhas reais;
- tokens e chaves;
- hostnames internos;
- endereços IP de produção;
- SQL contendo dados sensíveis;
- código-fonte proprietário de clientes ou da TOTVS.

O objetivo do projeto é permitir diagnóstico operacional com o menor nível de privilégio e exposição possível.
