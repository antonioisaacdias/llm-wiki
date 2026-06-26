---
id: truenas-longrunning-systemd-run
type: procedure
description: 'Jobs longos no TrueNAS devem rodar via `systemd-run` transiente, não tmux/nohup — KillUserProcesses mata processos de sessão SSH ao desconectar.'
tags: [truenas, systemd, ssh, jobs, background]
status: active
source: claude-code
created: 2026-06-25T00:00:00Z
modified: 2026-06-25T00:00:00Z
---
O TrueNAS tem `KillUserProcesses=yes` no logind, então processos iniciados numa sessão SSH (tmux, nohup, screen) são mortos quando a sessão desconecta. Para um job que precisa sobreviver ao logout, use uma unit transiente do systemd:

```
systemd-run --unit=meujob --description="..." /caminho/comando args
```

Acompanhe com `journalctl -u meujob -f` e pare com `systemctl stop meujob`. Veja também [[truenas-ix-apps-caps]] para gotchas de permissão no TrueNAS.
