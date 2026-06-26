---
id: ollama-thinking-toggle
type: procedure
description: 'No Ollama, desativar o chain-of-thought por request com `think: false` no body da API; o Modelfile não aceita esse parâmetro. Modelos como Gemma ficam muito mais rápidos em perguntas simples.'
tags: [ollama, thinking, cot, performance, api]
status: active
source: claude-code
created: 2026-06-25T00:00:00Z
modified: 2026-06-25T00:00:00Z
---
Para desligar o raciocínio (chain-of-thought) de um modelo no Ollama, passe `"think": false` no corpo JSON da requisição (`/api/chat` ou `/api/generate`). É por-request — o Modelfile NÃO aceita esse parâmetro, então não dá pra fixar via FROM/PARAMETER. Em perguntas simples isso acelera bastante modelos que pensam por padrão (ex.: Gemma grande chega a ~6x mais rápido). Mantenha o thinking ligado só quando a tarefa exige raciocínio.
