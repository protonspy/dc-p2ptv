---
status: accepted
---

# 0004 — Transporte por faixas de mídia WebRTC

## Context

O projeto antecessor desta ideia, `discord-screen-p2p`, entrega compartilhamento de tela
com **WebCodecs sobre WebSocket** e mede cerca de 40 ms de latência, sem ICE, sem STUN e
sem TURN. É um resultado real e o contraponto honesto a qualquer escolha feita aqui: a
complexidade do WebRTC foi evitada de propósito e funcionou.

O que aquele desenho não tem é o que este projeto precisa. WebSocket é TCP, então uma
perda bloqueia a fila inteira, e a recuperação de perda e o controle de congestionamento
precisariam ser escritos à mão. Não há mídia direta entre navegadores, porque não há ICE:
todo byte passa pelo servidor, que é justamente o custo que este projeto existe para
dividir. E cada espectador recebe a qualidade inteira, porque não há camadas — o número
que o próprio README registra, 2,5 Mb/s por espectador.

## Decision

O cliente usa **faixas de mídia WebRTC** para áudio, vídeo e tela. O navegador cuida da
codificação, do controle de congestionamento, da recuperação de perda e do simulcast; o
nó encaminha pacotes RTP sem abrir a carga.

ICE, STUN e TURN entram junto, como o preço de a mídia poder ir direto entre navegadores
na sala pequena.

## Consequences

Mídia direta entre navegadores passa a existir, que é o que dá nome ao projeto e o que
zera o custo de servidor nas salas de até quatro participantes.

Camadas de qualidade vêm do navegador, e o nó escolhe por assinante em vez de mandar a
qualidade inteira a todo mundo. É o segundo eixo de economia, ao lado da árvore.

TURN passa a ser necessário como rede de segurança, com custo próprio, e a fração de
sessões que passa por ele vira métrica de alarme.

A cifra tem de ser aplicada onde o quadro ainda é acessível ao código do cliente — no
transformador de fluxo codificado — porque a faixa de mídia, por si, seria decifrável
pelo nó que a encaminha.

WebCodecs continua sendo o caminho certo para quem precisa do quadro na mão. Se um
requisito futuro exigir isso, ele convive: WebCodecs sobre canal de dados, na mesma
sessão WebRTC, sem substituir este registro.
