# PPSPP — RFC 7574

O único protocolo de streaming P2P que virou padrão IETF (julho de 2015), saído do
grupo PPSP e da implementação `libswift` do projeto Tribler. Cobre sob demanda e ao
vivo com o mesmo conjunto de mensagens. Vale ler mesmo sem adotar: as decisões dele são
o vocabulário do problema descrito em [[p2ptv-fundamentos]].

## Identificação do enxame

O **swarm ID** é o conteúdo, não o servidor.

- **Sob demanda:** a raiz de uma árvore de Merkle sobre o conteúdo inteiro. Quem tem a
  raiz consegue verificar qualquer chunk vindo de qualquer par.
- **Ao vivo:** uma **chave pública**. Não existe raiz fixa porque a árvore ainda está
  crescendo.

## Endereçamento de chunks

Três esquemas; apenas o primeiro é obrigatório.

- **Chunk ranges** — início e fim, inteiros de 32 ou 64 bits.
- **Byte ranges** — deslocamentos de 64 bits, opcional.
- **Bin numbers** — numeração de intervalos binários sobre uma árvore balanceada
  mínima: a menor unidade é um chunk, a maior cobre 2^63 chunks, e o pai vale
  `(filhoEsq + filhoDir) / 2`. Um único inteiro endereça uma subárvore inteira, o que
  comprime o estado por par e encolhe o buffer map no fio. É a ideia mais reaproveitável
  do RFC.

## Mensagens

`HANDSHAKE` (aperto de três vias, IDs de canal aleatórios contra spoofing) · `HAVE`
(anuncia chunks já verificados) · `REQUEST` (puxa) · `DATA` · `ACK` (com atraso de via
única, para o controle de congestionamento) · `INTEGRITY` (hashes de verificação) ·
`PEX` (troca de pares) · `CHOKE` / `UNCHOKE`.

Um par só envia `HAVE` **depois de verificar** o chunk. É o que impede a propagação de
conteúdo poluído.

## Integridade

Folhas da árvore de Merkle são hashes de chunks; o pai é o hash da concatenação dos
filhos. O receptor recalcula até a raiz usando os **uncle hashes** que acompanham o
dado. O emissor evita reenviar hashes que o receptor já confirmou, e pode pular
otimisticamente hashes ainda não confirmados que já mandou.

**Ao vivo** a árvore cresce, então não há raiz para assinar. O injetor assina o
**munro hash** — a raiz de cada nova subárvore de `NCHUNKS_PER_SIG` chunks (potência de
2, mínimo 2). Uma assinatura amortizada sobre muitos chunks, verificável contra a chave
pública que é o swarm ID. Pares descartam o passado segundo uma **live discard window**
negociada no handshake.

## Canais e transporte

Vários enxames compartilham um endereço de transporte via **canais**: cada datagrama de
A para B leva na frente o ID de canal de 4 bytes que B alocou. Canal 0 é reservado ao
handshake.

Sobre UDP, com **LEDBAT** como controle de congestionamento — deliberadamente menos
agressivo que TCP, para ceder a passagem ao tráfego interativo do próprio usuário. Por
isso `DATA` carrega o relógio do emissor e `ACK` devolve o atraso de via única.

## Descoberta

Tracker centralizado, DHT ou PEX. O RFC não escolhe — é ponto de extensão explícito.
Para uso na Internet aberta, respostas PEX levam certificados de participação
assinados; em ambiente controlado, endereços puros bastam.

## Por que não pegou na web

Nada disso roda em browser: é UDP cru, e o browser só oferece WebRTC. Foi essa lacuna
que produziu o desenho de [[cdn-p2p-hibrido]], que reimplementa metade das ideias daqui
sobre DataChannel.
