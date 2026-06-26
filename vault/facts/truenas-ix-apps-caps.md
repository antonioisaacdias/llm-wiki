---
id: truenas-ix-apps-caps
type: fact
description: 'Apps ix-apps do TrueNAS rodam como root + CapDrop=ALL (sem CAP_DAC_OVERRIDE); mismatch de ownership no volume vira "readonly DB"/permission denied mesmo com o filesystem rw.'
tags: [truenas, ix-apps, docker, permissions, deploy]
status: active
source: claude-code
created: 2026-06-25T00:00:00Z
modified: 2026-06-25T00:00:00Z
---
Containers de apps ix-apps no TrueNAS SCALE rodam como root mas com todas as capabilities derrubadas (CapDrop=ALL), inclusive CAP_DAC_OVERRIDE. Consequência: se o dataset montado no container não for owned pelo UID que o processo usa, qualquer escrita falha com "readonly database" ou "permission denied" — mesmo o filesystem estando montado read-write. Mitigação: garantir que o dataset seja owned pelo UID do container (ou fixar PUID/PGID); nunca confiar só na flag rw do mount. Validar com um mount-test de escrita antes de declarar pronto.
