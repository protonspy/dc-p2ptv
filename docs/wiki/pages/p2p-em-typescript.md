# P2P de mídia em TypeScript

No navegador não se escolhe o transporte: é WebRTC ou nada. A escolha real é **o que
trafega** — quadros numa MediaStreamTrack, ou bytes num DataChannel — e essa decisão
define o sistema inteiro.

## Track de mídia contra DataChannel

**MediaStreamTrack.** O navegador cuida de codec, jitter buffer, NACK, FEC e
estimativa de banda. É o caminho para conversa e para tela ao vivo. Um par pode
**reencaminhar** uma track que recebeu para um terceiro, o que forma árvore de
distribuição no próprio navegador — mas cada salto soma latência e cada nó interno é um
espectador que pode fechar a aba. É o tree-push de [[p2ptv-fundamentos]] com todas as
fragilidades dele.

**DataChannel.** Bytes opacos, e portanto **segmentos**: fMP4 ou MPEG-TS numerados,
alimentados no player por MSE. É o que [[cdn-p2p-hibrido]] faz, e é o único caminho em
que um chunk recebido de um par é idêntico ao que veio da origem — condição para haver
enxame de verdade.

Regra prática: **conversa e tela interativa → track; transmissão para muitos →
DataChannel + segmentos.**

## Limites duros do DataChannel

O gargalo que mais surpreende. SCTP tem limite de 64 KB por mensagem, e o usrsctp do
Chromium **fecha o canal** com `EMSGSIZE` acima de 256 KiB.

- **Fatie em ~16 KiB.** Acima disso já é instável entre navegadores; acima de 64 KB é
  inviável.
- **Faça backpressure por `bufferedAmount`**, não por `setTimeout` — temporizador entre
  envios derruba a vazão. O padrão é chamar `send()` em laço até `bufferedAmount` passar
  do limiar, e retomar em `bufferedamountlow`.
- O buffer máximo por canal no Chrome é 16 MB; estourar fecha o canal.
- `bufferedAmount` não reflete fielmente o estado do buffer do SCTP — trate como
  aproximação.

## Bibliotecas

**Enxame de segmentos, pronto:**

- **`p2p-media-loader`** (Novage) — três pacotes: `-core`, `-hlsjs`, `-shaka`. Funciona
  com qualquer player construído sobre Hls.js ou Shaka (Vidstack, Clappr, Plyr,
  MediaElement, DPlayer, OpenPlayerJS). Ver [[player-de-streaming]].

  ```ts
  import Hls from "hls.js";
  import { HlsJsP2PEngine } from "p2p-media-loader-hlsjs";

  const HlsWithP2P = HlsJsP2PEngine.injectMixin(Hls);
  const hls = new HlsWithP2P({
    p2p: {
      core: {
        swarmId: "canal-42",              // default: URL do manifesto sem query
        segmentTtl: 60,
        httpDownloadTimeWindow: 3,        // orçamento de prazo, em segundos
        p2pDownloadTimeWindow: 6,
        simultaneousP2PDownloads: 3,
        validateP2PSegment: async (url, byteRange, data) => verificarHash(data),
        announceTrackers: ["wss://tracker.novage.com.ua"],
      },
      onHlsJsCreated: (hls) => {
        hls.p2pEngine.addEventListener("onPeerConnect", ({ peerId }) => {});
        hls.p2pEngine.addEventListener("onChunkDownloaded", () => {});
      },
    },
  });
  ```

  As duas janelas de tempo são o coração: `httpDownloadTimeWindow` é o quanto do buffer
  vem da origem por HTTP, `p2pDownloadTimeWindow` é o quanto pode esperar pelos pares.
  Mexer nelas é mexer diretamente na troca entre latência e descarga.
  `validateP2PSegment` é o gancho de integridade — sem ele, par malicioso envenena o
  enxame.

**Sinalização e sessão:**

- **`simple-peer`** — a camada fina clássica sobre `RTCPeerConnection`; ainda mantida.
- **PeerJS** — traz servidor de sinalização junto; bom para protótipo.
- **`trystero`** — sinalização "sem servidor" sobre Nostr (padrão), BitTorrent, MQTT,
  IPFS, Supabase ou Firebase, com salas e canais de dados prontos. Ressalva conhecida:
  não escala em número de pares por sala — o navegador barra com *Cannot create so many
  PeerConnections* — e o "sem servidor" depende de relays de terceiros.
- **`js-libp2p`** — o irmão do `go-libp2p` de [[p2p-em-go]]; mesma conclusão, serve à
  descoberta e não à mídia.

**Do lado do Node:**

- **`werift`** — WebRTC em TypeScript puro (ICE, DTLS, SCTP, RTP, SRTP, WebM, MP4).
  Útil quando o servidor precisa ser TypeScript.
- **`node-datachannel`** — binding de `libdatachannel` (C++); mais rápido, menos
  hackeável.
- **`webtorrent`** — BitTorrent que fala WebRTC e é compatível com os trackers que o
  `p2p-media-loader` usa para sinalizar.

## Teto de pares por página

Não existe enxame de milhares de vizinhos no navegador: o número de `RTCPeerConnection`
simultâneas é limitado, e cada uma custa handshake DTLS e memória. Dezenas de vizinhos
é o regime real, o que torna a **escolha de vizinho** (mesmo ASN, RTT baixo, quem tem
o segmento) mais importante do que o tamanho do enxame.

## Segurança

Todo par vê o IP de quem o serve, e o segmento vindo de um par é dado não confiável até
ser verificado. Pesquisa sobre entrega assistida por pares em WebRTC documenta o
levantamento de audiência por correlação de pares — o mesmo ponto de
[[cdn-p2p-hibrido]], e o motivo pelo qual o Discord recusa P2P
([[discord-arquitetura-de-midia]]).
