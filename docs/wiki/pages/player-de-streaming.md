# Player de streaming

O player é onde o P2P entra ou não entra. Se ele não expõe um gancho para trocar o
carregador de segmentos, não há onde plugar o enxame de [[p2p-em-typescript]].

## Duas famílias, e não são intercambiáveis

**MSE — segmentos.** `MediaSource` (e `ManagedMediaSource` no Safari do iOS 17+, o
subconjunto da Apple que finalmente deixa um player em JavaScript tocar DASH no iPhone)
recebe buffers de fMP4 ou MPEG-TS que o JavaScript entrega. Latência de 2 a 3 segundos
com LL-HLS ou LL-DASH bem ajustado. **É a única família que aceita P2P**, porque o
JavaScript vê e escolhe cada segmento.

**WebRTC — quadros.** `RTCPeerConnection` entrega a track direto ao elemento `<video>`.
Sub-segundo, jitter buffer e recuperação de perda por conta do navegador, e nenhum
ponto onde interceptar segmento — não existe segmento. Reprodução por WHEP cai aqui.

Escolher entre as duas é escolher entre **latência** e **possibilidade de enxame**. Não
há meio-termo, e é a decisão de arquitetura mais cara de reverter.

## Motores MSE

- **hls.js** — o player HLS de fato para navegador que não toca HLS nativamente. Tem
  `pLoader` e `fLoader` configuráveis: é exatamente aí que o `p2p-media-loader-hlsjs`
  se enfia, via `HlsJsP2PEngine.injectMixin(Hls)`.
- **Shaka Player** — do Google, fala HLS e DASH, e usa `ManagedMediaSource` no iOS.
  Integração P2P por `registerPlugins()` mais `bindShakaPlayer(player)`.
- **dash.js** — implementação de referência do DASH Industry Forum.
- **mpegts.js** — para MPEG-TS por WebSocket ou HTTP progressivo; latência baixa sem
  manifesto, ao custo de não ter bitrate adaptativo de verdade.

LL-HLS e LL-DASH empatam em latência prática, 2 a 3 segundos de vidro a vidro, e
empatam em incômodo operacional: partes parciais fazem o manifesto HLS ser reescrito a
cada 200 ms, e o CMAF em pedaços do DASH exige cooperação da CDN.

## Casca de interface

Separe motor de interface. O `hls.js` decodifica; quem desenha controles é outra coisa:

- **Vidstack** — componentes web e hooks, acessível, prefere `hls.js` ao motor nativo
  para ter comportamento igual em todo navegador, e permite apontar para a sua cópia de
  `hls.js` em vez da CDN padrão. É a opção moderna.
- **media-chrome** — só os controles, como componentes web; agnóstico de motor.
- **video.js** — o veterano; grande base de plugins, peso maior.

Qualquer um serve, desde que o motor por baixo seja `hls.js` ou Shaka — é o motor que o
enxame precisa alcançar, não a casca.

## O caminho novo: WebCodecs

`VideoEncoder` e `VideoDecoder` expõem codificação e decodificação quadro a quadro:
`VideoFrame` entra, `EncodedVideoChunk` sai, com a marcação de quadro-chave. Combinado
com **WebTransport** — que virou Baseline em março de 2026, com o Safari 26.4 — permite
montar a pilha inteira em JavaScript, sem MSE e sem WebRTC, com latência sub-segundo e
controle total do buffer.

É o que [[media-over-quic]] usa no navegador. Também é o gancho que devolve **objetos
endereçáveis** a um sistema de baixa latência, ou seja, reabre a porta para enxame P2P
sem pagar os 2 segundos do MSE. Ainda é montar tudo à mão: bitrate adaptativo, jitter
buffer e recuperação de perda não vêm de graça.

## Bitrate adaptativo e P2P brigam

O algoritmo de bitrate adaptativo mede a banda observada para decidir a qualidade. Com
P2P no meio, parte dos segmentos chega por um caminho que não representa a capacidade do
último quilômetro — o algoritmo lê rápido demais e sobe a qualidade além do sustentável,
ou o contrário. Ao integrar, verifique **qual número o algoritmo está enxergando** antes
de culpar o enxame por instabilidade de qualidade.
