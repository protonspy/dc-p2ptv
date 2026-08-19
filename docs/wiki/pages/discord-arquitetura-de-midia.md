# Discord — arquitetura de mídia

Discord **não** faz P2P. Toda voz e todo vídeo passam por um **SFU** (Selective
Forwarding Unit) próprio. A escolha é declarada e tem três razões: canais grandes com
malha completa custam `N²` conexões e uplink proibitivo; o servidor no meio esconde o
endereço IP dos participantes uns dos outros; e moderação (silenciar alguém de fato)
exige um ponto de controle no caminho da mídia.

Isso vale inclusive para chamada de duas pessoas em DM. A crença de que DM 1:1 é
direto é resíduo de outras plataformas VoIP — os IPs que aparecem em captura de rede
são os dos servidores do Discord.

## Plano de controle e plano de mídia

- **Gateway** — WebSocket persistente por cliente; presença, estado, eventos. Elixir.
- **Guilds** — dono do estado do servidor; observa a descoberta de serviço (`etcd`) e
  atribui ao guild o **voice server menos carregado da região**.
- **Voice server** — duas peças na mesma máquina: sinalização em **Elixir** e o **SFU em
  C++**, escrito em casa. O cliente abre um segundo WebSocket contra essa sinalização —
  é o voice gateway de [[discord-voice-gateway]].

Escala publicada na descrição da arquitetura: mais de 850 voice servers em 13 regiões e
mais de 30 datacenters, servindo 2,6 milhões de usuários de voz simultâneos, com mais de
220 Gbps e 120 Mpps de egresso.

## O cliente não é WebRTC padrão

No browser, é WebRTC comum: SDP, ICE, DTLS, SRTP.

No aplicativo nativo, é um media engine em C++ construído **sobre a biblioteca WebRTC**,
com o transporte trocado:

- **ICE fora.** Conexão direta ao servidor conhecido, mantida por pings periódicos. A
  sinalização de mídia cabe em cerca de 1000 bytes, contra o round-trip do browser.
- **DTLS/SRTP fora.** Cifra Salsa20 aplicada diretamente sobre a carga RTP, mais barata
  em CPU. Ver [[discord-transporte-de-midia]] para os modos atuais.
- **Acesso ao áudio cru**, o que permite detecção de atividade de voz do lado do cliente
  e **não transmitir nada durante o silêncio** — economia de banda e de CPU que importa
  porque, mesmo em canal enorme, só há poucos falantes simultâneos.

## Papel do SFU

Encaminha pacotes RTP sem transcodificar. Além disso: negocia o codec de envio de cada
cliente de modo que todos os demais consigam decodificar; processa RTCP (receiver
reports, perda, jitter) para otimização de qualidade; e comunica ao emissor, via *media
sink wants*, quais camadas e qualidades de vídeo alguém realmente quer receber.

No modo fim-a-fim ele **não consegue** decifrar os quadros — ver [[discord-dave-e2ee]].

## Falha e recuperação

- **SFU cai:** reinicia imediatamente e reconstrói o estado sem interação do cliente.
- **Voice server cai:** o cliente percebe a conexão cortada, pede verificação ao
  Gateway, e é reatribuído ao servidor menos carregado da região.
- **DDoS:** o servidor atingido é removido da descoberta de serviço; a reatribuição usa
  exatamente o mesmo caminho.

Vizinhança conceitual: [[sfu-simulcast-e-svc]] compara esse desenho com Meet, Jitsi e
LiveKit.
