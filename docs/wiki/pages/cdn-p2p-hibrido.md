# CDN + P2P híbrido no browser

O que sobrou de [[p2ptv-fundamentos]] na web comercial: nenhum sistema sério é P2P puro.
A CDN continua sendo a fonte e o piso de qualidade; o enxame é uma camada de desconto
sobre ela. Peer5, Streamroot (hoje Lumen), CDNBye e o open-source P2P Media Loader
seguem todos esse molde.

## Como funciona

O fluxo já é HLS ou DASH — segmentos numerados, servidos por HTTP. O motor P2P se
enfia entre o player e a rede, substituindo o carregador de segmentos:

1. O gerenciador de buffer do player pede os próximos N segmentos.
2. O escalonador P2P verifica quais pares já têm cada segmento e conseguem entregá-lo
   dentro do orçamento de latência — na ordem de 200 ms ao vivo, alguns segundos sob
   demanda.
3. O que couber, vem do enxame por **WebRTC DataChannel** (SCTP sobre DTLS).
4. O que não couber no prazo, cai para a borda da CDN.

Esse recuo por prazo é a diferença entre híbrido e BitTorrent ingênuo: o P2P nunca
segura o player. Alguns pares sorteados baixam segmentos novos por HTTP e os semeiam —
é assim que o conteúdo entra no enxame antes de qualquer um o ter.

## Sinalização e formação do enxame

O par precisa de um canal fora de banda para trocar SDP e candidatos ICE. P2P Media
Loader 1.x usa trackers compatíveis com WebTorrent; a série 2.x foi para uma abordagem
sem tracker, inspirada em DHT. Os fornecedores comerciais mantêm clusters de tracker
próprios, porque o tracker é onde se implementa a política que decide a qualidade:
agrupar pares por ASN, por geografia e por RTT medido.

## Números realistas

Descarga relatada em implantações bem ajustadas (2026): **60–80%** em eventos ao vivo
grandes (100 mil ou mais simultâneos), **30–55%** em ao vivo de porte médio, **5–15%**
em VOD de cauda longa. Estudos acadêmicos otimizados chegam a 88% dos chunks via P2P
com CDN em 12% e latência média de 148 ms. Métricas de qualidade ficam tipicamente
dentro de 10–15% do resultado só-CDN.

A variável dominante é **densidade de pares por segmento**: muita gente assistindo a
mesma coisa ao mesmo tempo, na mesma região. Público disperso, catálogo longo ou
audiência majoritariamente móvel derrubam a descarga — NAT de operadora (CGNAT) impede
conexão direta, e mais de 90% dos pares em sistemas de streaming estão atrás de algum
NAT.

## Limites duros do browser

- **DataChannel** é SCTP sobre DTLS: confiável e ordenado por padrão, com custo de
  cabeçalho e de handshake por par. Segmentos grandes precisam ser fatiados.
- **Travessia de NAT** exige STUN, e o que não passa exige TURN — que é um relay pago,
  ou seja, exatamente o custo que o P2P queria evitar. TURN é rede de segurança, não
  caminho de dados.
- **Número de conexões** por página é limitado; o enxame de um par é dezenas de
  vizinhos, não milhares.
- **Exposição de IP entre pares** é inerente. Trabalho acadêmico sobre entrega assistida
  por pares em WebRTC documenta esse e outros riscos — pares podem ser censados e
  correlacionados com o que assistem.

## O que isso ensina para transmissão sobre plataformas fechadas

O padrão que funciona é sempre o mesmo: **origem confiável + enxame oportunista +
recuo por prazo**. As peças concretas para montar isso estão em [[p2p-em-go]],
[[p2p-em-typescript]] e [[player-de-streaming]], e o encaixe delas em
[[arquitetura-de-referencia]]. Onde a origem é um SFU que não expõe os segmentos como objetos
endereçáveis — o caso do Discord, ver [[discord-transporte-de-midia]] — a primeira
metade do desenho não existe, e é isso que [[viabilidade-p2ptv-sobre-discord]] discute.
