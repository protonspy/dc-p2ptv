# Arquitetura de referência

Como as peças de [[p2p-em-go]], [[p2p-em-typescript]], [[player-de-streaming]] e
[[captura-e-transmissao-de-tela]] se encaixam. Não é decisão tomada — é o desenho que a
pesquisa sustenta, para ser aceito ou recusado com argumento.

## A bifurcação, antes de qualquer código

**Qual latência o produto precisa?** A resposta separa dois sistemas que quase não
compartilham peça.

| | Interativo | Difusão |
|---|---|---|
| Latência | sub-segundo | 2 a 5 segundos |
| Transporte | WebRTC, track de mídia | HTTP + segmentos, MSE |
| Fan-out | SFU (servidor) | CDN ou origem + enxame P2P |
| P2P entre espectadores | não cabe | é o ponto |
| Custo por espectador | linear no servidor | cai com a densidade do enxame |

Querer os dois ao mesmo tempo é o erro caro. **Se há chat de voz ao lado, são dois
caminhos**: voz e tela do apresentador em WebRTC para quem participa, e a mesma tela
empacotada em segmentos para quem só assiste.

## Desenho de difusão, ponta a ponta

```
captura (tela/câmera)
   -> ingestão WHIP  ->  [Go: origem]  ->  fMP4 + manifesto LL-HLS
                              |
                              +--> CDN ou HTTP próprio  --> navegador
                                                             |
                    tracker de sinalização (Go) <------------+
                                                             |
                                        enxame WebRTC DataChannel entre pares
```

**Ingestão.** WHIP, que é RFC 9725 desde março de 2025: um `POST` com SDP. OBS, FFmpeg,
GStreamer e navegador já falam, então não se inventa protocolo de sinalização.

**Origem em Go.** `pion/webrtc` termina a sessão WHIP; `gohlslib` empacota em fMP4 e
escreve o manifesto de baixa latência. Sem transcodificação — o codec que entra é o que
sai, que é a razão de o custo por espectador ser plano. Enquanto a origem própria não
existe, `mediamtx` cobre o papel inteiro e fala WHIP, WHEP, LL-HLS e SRT.

**Tracker em Go.** WebSocket, compatível com o protocolo de tracker WebTorrent, para que
o `p2p-media-loader` funcione sem adaptação. É aqui que mora a política que decide a
descarga: agrupar pares por ASN e por região, limitar o tamanho da vizinhança, e nunca
por peso de CPU — é sinalização, não mídia.

**Navegador.** `hls.js` como motor, Vidstack como casca, `p2p-media-loader-hlsjs` entre
os dois. A configuração que importa é `httpDownloadTimeWindow` contra
`p2pDownloadTimeWindow`: o primeiro é o pedaço do buffer que vem da origem e garante que
o player nunca trave, o segundo é a folga concedida ao enxame.

**TURN.** `pion/turn`, dimensionado como rede de segurança e não como caminho de dados.
Se o tráfego de TURN crescer, o P2P está pagando o custo que veio eliminar.

## Desenho interativo, ponta a ponta

SFU próprio em Go sobre Pion — o esqueleto está em [[p2p-em-go]], com a advertência
sobre drenar RTCP — ou LiveKit, que já traz o cluster e a malha distribuída de
[[sfu-simulcast-e-svc]]. Cliente por WHEP ou pelo SDK.

Escrever o SFU se paga quando o roteamento é a regra de negócio; não se paga para ter
uma sala de vídeo. Nesse caso o LiveKit resolve, e o esforço vai para o que é próprio.

## Onde P2P entre espectadores realmente compensa

Números de [[cdn-p2p-hibrido]]: 60 a 80% de descarga só aparece com audiência grande,
concentrada no tempo e na geografia. Antes disso o enxame é complexidade sem retorno.

O gatilho honesto é **densidade de pares por segmento**, não total de usuários. Vale
instrumentar a taxa de descarga desde o primeiro dia e desligar o enxame se ela não
sair do chão — desligar é uma linha de configuração, e a decisão fica com o dado.

## Sequência de construção

1. Origem e reprodução sem P2P nenhum. WHIP para dentro, LL-HLS para fora, player
   tocando. É o piso de qualidade e a referência contra a qual tudo se mede.
2. Instrumentação: latência de vidro a vidro, travamentos, tempo até o primeiro quadro.
3. Tracker e enxame, com recuo incondicional para a origem por prazo.
4. Medir a descarga contra o piso do passo 1. Só então ajustar as janelas de tempo.
5. Integridade de segmento (`validateP2PSegment`) antes de qualquer exposição pública —
   sem ela, um par malicioso envenena o enxame, que é o problema que o
   [[ppspp-rfc7574]] resolve com Merkle.

Nada disso depende do Discord, e é de propósito: ver [[viabilidade-p2ptv-sobre-discord]]
para o que a plataforma permite e o que não permite.
