# Security

Protheus MCP is designed as a read-only diagnostic bridge between MCP-compatible AI clients and ERP infrastructure.

## v0.1 security guarantees

The project does not intentionally expose tools for:

- executing arbitrary SQL;
- modifying ERP tables;
- killing database sessions;
- terminating processes;
- restarting Windows services;
- editing configuration files;
- returning configured database passwords.

Use a dedicated database login with the minimum permissions required for monitoring DMVs. Do not reuse privileged application or administrator credentials.

## Reporting a vulnerability

Please avoid publishing credentials, production logs, database addresses, or confidential ERP information in public issues. Report security problems privately to the maintainer whenever possible.
