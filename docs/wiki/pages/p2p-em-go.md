# P2P de mídia em Go

Go não tem WebRTC nativo do navegador para copiar: tem **Pion**, uma implementação
pura em Go de toda a pilha. É o que torna a linguagem viável aqui — e é a base de
quase tudo que existe no ecossistema.

## A pilha Pion

Módulos separados, todos usáveis isoladamente:

| Módulo | O que resolve |
|---|---|
| `pion/webrtc` | a API WebRTC completa: PeerConnection, tracks, DataChannel |
| `pion/ice`, `pion/turn`, `pion/stun` | travessia de NAT, e um servidor TURN próprio |
| `pion/dtls`, `pion/srtp` | DTLS 1.2 e cifra de mídia |
| `pion/sctp`, `pion/datachannel` | o transporte de dados do WebRTC |
| `pion/rtp`, `pion/rtcp` | pacotes e feedback |
| `pion/interceptor` | NACK, relatórios de emissor/receptor, TWCC, jitter buffer |
| `pion/mediadevices` | captura de câmera, microfone e tela — ver [[captura-e-transmissao-de-tela]] |

Suporta simulcast, SVC, NACK, sender/receiver reports e TWCC — o vocabulário de
[[sfu-simulcast-e-svc]] inteiro.

## Interceptors

O conceito central da v3 em diante. Uma cadeia plugável que processa RTP e RTCP na
entrada e na saída, sem que o código de aplicação toque nos pacotes:

```go
m := &webrtc.MediaEngine{}
i := &interceptor.Registry{}
webrtc.RegisterDefaultInterceptors(m, i) // NACK, reports, TWCC
webrtc.ConfigureNack(m, i)
webrtc.ConfigureTWCCHeaderExtensionSender(m, i)
webrtc.ConfigureSimulcastExtensionHeaders(m)
api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i))
```

## Encaminhar mídia — o esqueleto de um SFU

Uma track pode ser adicionada a **várias** PeerConnections; é isso que torna o SFU
simples em Pion. O padrão é: receber em `OnTrack`, escrever em `TrackLocalStaticRTP`
que já foi adicionada aos assinantes.

**A armadilha:** é obrigatório drenar o RTCP dos dois lados. Sem os laços de leitura,
NACK, PLI e relatórios são silenciosamente descartados — a chamada "funciona" e a
qualidade não se recupera de perda nenhuma.

```go
// publicador
peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
    go func() {
        for {
            if _, _, err := receiver.ReadRTCP(); err != nil {
                return // io.EOF encerra
            }
        }
    }()
    // ... encaminhar RTP para as tracks locais dos assinantes
})

// assinante — drenar o RTCP de cada sender
go func() {
    buf := make([]byte, 1500)
    for {
        if _, _, err := rtpSender.Read(buf); err != nil {
            return
        }
    }
}()
```

O interceptor `intervalpli` cobre o pedido periódico de quadro-chave, sem o qual quem
entra no meio da transmissão fica na tela preta até o próximo keyframe natural.

`TrackLocalStaticRTP.WriteRTP` copia o pacote de um `sync.Pool` a cada chamada — é onde
o custo aparece quando o fan-out cresce.

## O que já vem pronto

- **LiveKit** — SFU em Go sobre Pion, Apache-2.0, com plataforma inteira em volta.
  Cluster de nós idênticos coordenados por Redis, e a malha distribuída de
  [[sfu-simulcast-e-svc]] como comportamento padrão. É o que tem tração hoje.
- **Galene** — SFU pequeno em Go, desde a 1.0 sobre Pion v4, com multiplexação de UDP.
- **ion-sfu** — também sobre Pion; manutenção incerta, verificar antes de adotar.
- **MediaMTX** (`bluenviron/mediamtx`) — servidor sem dependências que fala WHIP, WHEP,
  RTMP, SRT, RTSP, LL-HLS e Media over QUIC, e converte entre eles. Serve como origem
  pronta enquanto o P2P é desenvolvido.
- **gohlslib** (`bluenviron/gohlslib`) — cliente e muxer HLS, com escrita de fMP4 e modo
  de baixa latência. É a peça que produz os segmentos que um enxame vai trocar.

## WHIP e WHEP

`WHIP` virou RFC 9725 em março de 2025: publicação WebRTC reduzida a um `POST` HTTP com
SDP no corpo. `WHEP` é o espelho para reprodução. Elimina a invenção de protocolo de
sinalização — vale adotar mesmo num sistema fechado, porque FFmpeg, GStreamer, OBS e
navegadores já falam.

## Fora do WebRTC

- **`quic-go`** e **`quic-go/webtransport-go`** — o caminho para
  [[media-over-quic]] e para WebTransport no navegador.
- **`go-libp2p`** com **gossipsub** — overlay P2P pronto, com descoberta por DHT.
  Cuidado: gossipsub é desenhado para mensagens, não para mídia. Medições mostram a
  latência degradando para ~4 s sob 20 mensagens/s, e a v2.0 troca duplicatas por mais
  latência. O `episub`, extensão para multicast de fonte única, existe justamente porque
  o gossipsub base não serve a esse caso. Serve para **sinalização e descoberta**, não
  para carregar quadros.
- **`anacrolix/torrent`** — BitTorrent maduro em Go, útil para VOD, não para ao vivo.

## Onde Go se paga

Terminação de milhares de conexões UDP com goroutines baratas, e um SFU que é código
seu em vez de configuração de servidor alheio. Onde não se paga: codificação e
decodificação de vídeo. Não há encoder puro em Go que preste; qualquer transcodificação
significa cgo com libvpx, x264, libaom ou GStreamer — e a decisão certa quase sempre é
**não transcodificar**, que é exatamente o argumento do SFU.

Contraparte no navegador: [[p2p-em-typescript]].
