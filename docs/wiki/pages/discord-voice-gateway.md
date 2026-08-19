# Discord — voice gateway

O plano de controle da sessão de mídia: um WebSocket separado do gateway principal,
contra o voice server atribuído pelo Guilds ([[discord-arquitetura-de-midia]]). Versão
passada em `?v=`; v8 introduziu resume com buffer, e vídeo funciona a partir da v5.
Payloads JSON vão em text frames — binary frame só para o subprotocolo DAVE.

## Sequência de conexão

1. **Op 0 Identify** — `server_id`, `channel_id`, `user_id`, `session_id`, `token`, e a
   versão máxima de DAVE suportada.
2. **Op 2 Ready** — SSRC atribuído, IP e porta UDP do servidor, lista de modos de cifra
   disponíveis, streams negociados.
3. **IP discovery** — o cliente descobre seu IP e porta externos mandando um datagrama
   ao endereço recebido:

   | Campo | Tipo | Tamanho |
   |---|---|---|
   | Type | unsigned short, big endian | 2 |
   | Length | unsigned short, big endian | 2 |
   | SSRC | unsigned int, big endian | 4 |
   | Address | string terminada em nulo | 64 |
   | Port | unsigned short, big endian | 2 |

4. **Op 1 Select Protocol** — anuncia `udp` ou `webrtc`, o IP e porta descobertos, o
   modo de cifra escolhido e a lista de codecs.
5. **Op 4 Session Description** — devolve `audio_codec`, `video_codec`, o `secret_key`
   da sessão e o modo confirmado.
6. **Op 5 Speaking** — obrigatório **antes** de mandar ou receber mídia. Sem isso a
   conexão cai com *Invalid SSRC*.

Flags de speaking: `1<<0` VOICE, `1<<1` SOUNDSHARE (áudio de contexto do vídeo),
`1<<2` PRIORITY. Também viajam em extensão RTP de ID 9, comprimidas em um byte:
`((flags & 0x03) << 1) | ((flags & 0x04) >> 2)`.

## Resume com buffer (v8+)

O cliente guarda o número de sequência da última mensagem recebida e o devolve como
`seq_ack` em heartbeats e no resume. Na volta, o gateway reenvia o que ficou no buffer.
Mensagens binárias servidor→cliente levam sequência de 2 bytes e opcode de 1 byte antes
da carga; no sentido inverso a sequência é omitida.

## Opcodes DAVE

`21` prepare transition (rebaixar para v0) · `22` execute transition · `23` transition
ready · `24` prepare epoch · `25` external sender package · `26` key package · `27`
proposals · `28` commit welcome · `29` announce commit transition · `30` welcome · `31`
invalid commit welcome. Semântica em [[discord-dave-e2ee]].

## Sinalização de Go Live

Compartilhamento de tela não reusa o canal de voz: o cliente manda `STREAM_CREATE` e
`STREAM_SET_PAUSED` no gateway principal, e recebe `STREAM_SERVER_UPDATE` com **outro
endpoint de voice gateway**, dedicado ao stream. Antes de emitir vídeo, manda ainda o
payload de vídeo com resolução, taxa de quadros e bitrate desejados. O envio só começa
quando há alguém assistindo; áudio, ao contrário, continua saindo mesmo em canal vazio.

O formato dos pacotes está em [[discord-transporte-de-midia]].
