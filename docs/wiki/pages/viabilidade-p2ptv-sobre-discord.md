# Viabilidade de P2PTV sobre Discord

Síntese das outras páginas, aplicada à pergunta que dá nome a este repositório: dá para
distribuir por pares um fluxo que nasce no Discord?

## As quatro barreiras, em ordem de dureza

**1. Não há objeto para trocar.** P2P de streaming — [[p2ptv-fundamentos]],
[[ppspp-rfc7574]], [[cdn-p2p-hibrido]] — pressupõe conteúdo fatiado em unidades
nomeadas e idênticas para todo mundo: chunk, segmento, objeto. O que o Discord entrega
é RTP cifrado por sessão ([[discord-transporte-de-midia]]). Dois espectadores recebem
bytes diferentes, com chaves diferentes, sem numeração compartilhada. Não existe "o
segmento 4821" para pedir a um vizinho.

**2. A chave é por sessão, e sob DAVE nem o servidor tem.** A cifra de transporte usa o
`secret_key` do Session Description daquele cliente. Sob [[discord-dave-e2ee]], a mídia
é cifrada quadro a quadro com chave por emissor derivada do grupo MLS, e só membros do
grupo decifram. Repassar bytes crus a um não-membro repassa ruído.

**3. Termos de uso.** Os termos do Discord proíbem engenharia reversa e software não
autorizado que modifique o serviço; automatizar conta de usuário fora da API oficial de
bots ("self-bot") é proibido desde 2017 e é motivo de encerramento de conta. A API de
bots **não** dá recepção de mídia de voz de forma suportada. Qualquer desenho que
dependa de um cliente não oficial entrando no canal para capturar o fluxo é violação
contratual, com risco de banimento — e vale dizer isto antes de escrever código, não
depois.

**4. Contradição de propósito.** O Discord usa SFU justamente para não expor IP entre
participantes ([[discord-arquitetura-de-midia]]). Um enxame P2P entre espectadores
reintroduz exatamente essa exposição.

## O que resta, e é bastante

As barreiras acima matam "P2P *dentro* do Discord". Não matam duas coisas legítimas:

**A. Discord como plano de controle, mídia própria.** O Discord vira sinalização,
identidade e presença — que é onde a plataforma é boa e onde a API oficial existe —
enquanto o fluxo de vídeo é seu: origem própria em HLS/LL-HLS ou WebRTC, e enxame P2P
por WebRTC DataChannel entre os espectadores, no molde de [[cdn-p2p-hibrido]]. Todo o
desenho maduro está disponível: segmentos numerados, escalonador com prazo, recuo para
a origem, tracker agrupando por ASN e RTT. O desenho ponta a ponta, com as bibliotecas
nomeadas, está em [[arquitetura-de-referencia]].

**B. Um SFU próprio, com distribuição em árvore de servidores.** Se o objetivo é o
fan-out barato e não o P2P por si, [[sfu-simulcast-e-svc]] descreve a alternativa que a
indústria escolheu: cascata de SFUs (Octo, LiveKit) entrega distribuição geográfica sem
expor pares e sem depender de uplink doméstico. E [[media-over-quic]] é para onde isso
migra quando a pilha amadurecer.

## Se for o caminho A, o que decidir cedo

- **Tamanho do segmento** é o botão que troca latência por eficiência de enxame.
  Segmento curto baixa o atraso e piora a chance de um vizinho já o ter.
- **Orçamento de prazo** por segmento, com recuo incondicional para a origem. Sem isso,
  o P2P trava o player, que é o erro clássico.
- **Sinalização**: tracker próprio ou compatível com WebTorrent. É onde se implementa
  agrupamento por ASN e região, e é o que decide a taxa de descarga.
- **TURN**: obrigatório como rede de segurança, e é custo — nunca caminho de dados.
- **Integridade dos chunks**: sem verificação, um par malicioso envenena o enxame. A
  ideia de Merkle do [[ppspp-rfc7574]] resolve isso e é barata de reimplementar sobre
  segmentos HLS.
- **Expectativa de descarga**: 60–80% só aparece com audiência grande, concentrada e
  simultânea. Público pequeno ou disperso entrega pouco, e o P2P vira complexidade sem
  retorno.

Se o fluxo for tela em vez de câmera, [[captura-e-transmissao-de-tela]] muda as escolhas
de codec e de captura antes de qualquer decisão de rede.
