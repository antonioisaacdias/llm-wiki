---
id: kopia-snapshot-list-collapses-identical
type: fact
description: '`kopia snapshot list` colapsa snapshots idênticos por padrão; use `--show-identical`/`--json` pra ver o último backup real, senão widgets de monitoramento dão falso-positivo de atraso.'
tags: [kopia, backup, monitoring, snapshots]
status: active
source: claude-code
created: 2026-06-25T00:00:00Z
modified: 2026-06-25T00:00:00Z
---
Por padrão o `kopia snapshot list` agrupa/colapsa snapshots de conteúdo idêntico, mostrando só o primeiro. Para sources estáveis (que mudam pouco), isso faz parecer que o último backup é antigo, gerando falso-positivo de "backup atrasado" em widgets de dashboard. Use `--show-identical` ou `--json` para enxergar o snapshot mais recente de verdade ao calcular a idade do último backup.
