# Discord — transporte de mídia

UDP com RTP, negociado por [[discord-voice-gateway]]. Não é SRTP: a cifra é aplicada
pelo próprio Discord sobre a carga, com o cabeçalho RTP em claro para que o SFU possa
encaminhar.

## Cabeçalho RTP

| Campo | Tipo | Tamanho |
|---|---|---|
| Version + flags | unsigned byte | 1 |
| Payload type | unsigned byte | 1 |
| Sequence | unsigned short, big endian | 2 |
| Timestamp | unsigned int, big endian | 4 |
| SSRC | unsigned int, big endian | 4 |
| CSRCs | opcional | n |
| Extension | opcional | n |
| Payload | | n |

Áudio Opus: relógio de 48 kHz, quadro de 20 ms, timestamp avança 960 amostras por
quadro. Vídeo: relógio de 90 kHz, bit *marker* no último pacote de cada quadro.

## Modos de cifra

| Modo | Situação |
|---|---|
| `aead_aes256_gcm_rtpsize` | preferir quando disponível |
| `aead_xchacha20_poly1305_rtpsize` | obrigatório suportar |
| `xsalsa20_poly1305_lite_rtpsize` | descontinuado |

O AES-GCM depende de aceleração no hardware do servidor, então nem sempre é ofertado;
o XChaCha20 está sempre na lista. Ambos usam nonce de inteiro incremental de 32 bits
**anexado ao fim do datagrama** — quem decifra tira o nonce do fim antes de processar.
Nos modos `_rtpsize`, a fronteira do que fica em claro segue a mesma regra do SRTP:
CSRCs e, opcionalmente, o preâmbulo da extensão contam como cabeçalho não cifrado.

A chave é o `secret_key` entregue no Session Description. É cifra **de transporte**:
o SFU tem a chave e vê a mídia. Confidencialidade contra o próprio servidor é outra
camada — [[discord-dave-e2ee]].

## SSRC e retransmissão

O Ready atribui o SSRC de áudio e os SSRCs de vídeo, em pares primário + RTX por camada
de simulcast; quando o RTX é omitido, vale `primário + 1`. Pacotes primários usam o
payload type do codec negociado e o RID da camada; pacotes RTX usam o payload type de
RTX e a extensão de RID reparado, com o número de sequência original no começo da carga
(RFC 4588). O receptor remapeia o RTX para o SSRC primário antes de despacotizar.

Duas portas UDP no cliente nativo: uma para mídia, outra para RTX.

## Codecs

Áudio: **Opus**, obrigatório, prioridade máxima.

Vídeo, por ordem de preferência: **AV1**, **H.265**, **H.264** (padrão), **VP8**, **VP9**.
Cada um declara `payload_type` e `rtx_payload_type`. **Payload type 96 é reservado** a
pacotes de sondagem de banda e não pode ser usado por codec.

Capturas de tráfego antigas mostram 120 para Opus, 101 para H.264 e 102 para o RTX
correspondente — mas isso é negociado, não fixo.

## Qualidade de vídeo

Go Live usa o encoder de hardware quando existe. Nível gratuito vai a 720p30; Nitro
libera 1080p60 e acima, com bitrate na casa de 6 Mbps para 1080p60. O SFU escolhe qual
camada encaminhar a cada espectador segundo os *media sink wants* dele, o que é o
mecanismo padrão de simulcast descrito em [[sfu-simulcast-e-svc]].

Limite típico de 50 usuários com vídeo por canal de voz; canais de palco têm limite
próprio.

## O que isso significa para redistribuição

Não existe segmento endereçável aqui. O que passa no fio é um fluxo RTP cifrado por
sessão, com chave por sessão, sem numeração estável de conteúdo compartilhável entre
espectadores. É a diferença estrutural com HLS que [[viabilidade-p2ptv-sobre-discord]]
tem de resolver.
