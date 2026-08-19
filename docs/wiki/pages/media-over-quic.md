# Media over QUIC

O grupo de trabalho MoQ do IETF, documento principal `draft-ietf-moq-transport`. Em
março de 2026 estava na revisão -17, com `-latest` em maio de 2026; **ainda não é RFC**.

A proposta é ocupar o vão entre os dois mundos descritos em [[cdn-p2p-hibrido]] e
[[sfu-simulcast-e-svc]]: escalabilidade de árvore de CDN com latência sub-segundo que
hoje só WebRTC entrega. Publicador manda para um **relay**; relays formam árvore e
replicam para assinantes. Sem SDP, sem ICE, sem SRTP — QUIC resolve multiplexação,
cifra e controle de congestionamento, e o browser chega via WebTransport.

O que muda de relevante para quem pensa em distribuição:

- **Objetos endereçáveis de novo.** MoQ nomeia tracks, grupos e objetos. Isso devolve
  a propriedade que HLS tem e que RTP não tem — a mesma propriedade que faz enxame P2P
  ser possível sobre HLS e impossível sobre o transporte do Discord
  ([[discord-transporte-de-midia]]).
- **Relay é cacheável e encadeável**, então a árvore de distribuição é do protocolo, não
  de um produto.

Estado prático em 2026: onze fornecedores (Ant Media, AWS, Bitmovin, Broadpeak,
CacheFly, Cloudflare, Nomad Media, Oracle, Norsk, Synamedia, Red5) demonstraram
implementações no NAB 2026, e demos híbridas mostram publicador WebRTC alimentando
árvore de relays MoQ com tradução na borda. Ainda assim: WebTransport é fração de
percentual dos carregamentos de página, contra ~0,35% de WebRTC. Não é escolha de
produção hoje; é a direção para onde o campo vai.
