---
id: vram-overflow-cliff
type: fact
description: 'Em Ollama, overflow de VRAM com poucas layers offloaded deixa a inferência ~4x mais lenta; "quase coube" é pior que um modelo menor que coube inteiro.'
tags: [ollama, vram, performance, gpu, inference]
status: active
source: claude-code
created: 2026-06-25T00:00:00Z
modified: 2026-06-25T00:00:00Z
---
Quando um modelo no Ollama não cabe inteiro na VRAM e só algumas layers (ex.: 6) são offloaded pra GPU, o resto roda na CPU e a inferência fica cerca de 4x mais lenta. O penhasco é abrupto: um modelo que "quase coube" performa pior que um modelo menor que coube 100% na GPU. Ao escolher quantização/tamanho, prefira o que cabe inteiro a maximizar parâmetros.
