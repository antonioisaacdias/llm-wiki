---
id: node-exporter-textfile-perms
type: fact
description: 'O textfile collector do node-exporter exige .prom com permissão 0644; mktemp cria 0600 e o user prometheus não lê → chmod 0644 antes do mv. Sintoma: node_textfile_scrape_error 1.'
tags: [prometheus, node-exporter, monitoring, permissions]
status: active
source: claude-code
created: 2026-06-25T00:00:00Z
modified: 2026-06-25T00:00:00Z
---
Ao escrever métricas custom pro textfile collector do node-exporter, o arquivo `.prom` precisa ser legível pelo user `prometheus` (modo 0644). `mktemp` cria com 0600, então se você gera num temp e dá `mv` pro diretório do collector, o node-exporter não consegue ler e expõe `node_textfile_scrape_error 1`. Faça `chmod 0644` no arquivo antes do `mv` atômico para o diretório final.
