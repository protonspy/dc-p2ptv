# Wiki

O ponto de entrada. Toda página em `wiki/pages/` precisa ser alcançável daqui — direta
ou indiretamente — porque página que ninguém liga é página que ninguém acha de novo.

Páginas se ligam como `[[page-slug]]`, onde o slug é o nome do arquivo sem extensão e
sem diretório.

## Páginas

### Distribuição P2P de vídeo

- [[p2ptv-fundamentos]] — o problema, tree-push contra mesh-pull, escalonamento de
  chunks, incentivo, métricas
- [[ppspp-rfc7574]] — o padrão IETF de streaming P2P: bin numbers, munro hashes, LEDBAT
- [[cdn-p2p-hibrido]] — o que sobrou disso na web: enxame WebRTC como desconto sobre CDN

### Conferência em tempo real

- [[sfu-simulcast-e-svc]] — malha contra MCU contra SFU, simulcast contra SVC, SFU em
  cascata, Google Meet
- [[media-over-quic]] — para onde a entrega de baixa latência está indo

### Implementação

- [[p2p-em-go]] — a pilha Pion, interceptors, o esqueleto de um SFU, LiveKit, MediaMTX
- [[p2p-em-typescript]] — track contra DataChannel, limites do SCTP, p2p-media-loader
- [[player-de-streaming]] — MSE contra WebRTC, hls.js e Shaka, WebCodecs, onde o P2P
  se pluga
- [[captura-e-transmissao-de-tela]] — getDisplayMedia, contentHint, áudio do sistema,
  codec para texto, captura em Go
- [[arquitetura-de-referencia]] — como as peças se encaixam, e em que ordem construir

### Discord

- [[discord-arquitetura-de-midia]] — SFU próprio, Elixir e C++, escala, falha, por que
  não P2P
- [[discord-voice-gateway]] — handshake, opcodes, IP discovery, Go Live
- [[discord-transporte-de-midia]] — RTP, SSRC e RTX, modos de cifra, codecs
- [[discord-dave-e2ee]] — MLS, epochs, formato do quadro cifrado, passthrough

### Síntese

- [[viabilidade-p2ptv-sobre-discord]] — o que é possível, o que é proibido, o que decidir
