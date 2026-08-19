# SFU, simulcast e SVC

As três topologias de conferência, e por que uma venceu.

- **Malha (mesh) P2P.** Cada um envia para cada um: `N-1` uploads por participante,
  `N²` fluxos. Funciona até 3 ou 4 pessoas e morre no uplink depois disso.
- **MCU.** O servidor decodifica tudo, compõe um mosaico e reencoda. Um fluxo por
  cliente, mas CPU proibitiva e latência de transcodificação.
- **SFU.** O servidor **encaminha pacotes sem reencodar**. Um upload por participante,
  seleção por receptor no servidor. Barato o bastante para escalar e simples o bastante
  para ser confiável. É o que Discord ([[discord-arquitetura-de-midia]]), Meet, Jitsi,
  LiveKit, mediasoup e Janus usam.

## Como o SFU decide o que mandar

Ele não pode reencodar, então precisa que o emissor já produza opções.

**Simulcast** — o emissor codifica a mesma câmera em várias resoluções e bitrates
independentes e sobe todas. O SFU escolhe uma por receptor. Custo: encoder múltiplo no
emissor e upload somado das camadas.

**SVC** — um único fluxo com camadas embutidas, temporais (taxa de quadros) e espaciais
(resolução). O SFU descarta camadas para baixar a qualidade. Mais eficiente em banda de
subida; exige codec que suporte (VP9, AV1, H.265) e, na prática, decodificação por
software, porque decoder de hardware frequentemente não dá conta de SVC.

Google Meet roda **VP9 K-SVC** em produção com modo `L3T3_KEY`, e liga e desliga
conforme o número de participantes — `L1T1`, sem SVC, no aquecimento antes da chamada,
migrando para `L3T3_KEY` conforme gente entra. **AV1** já está em produção em Meet, Teams
e Messenger, mas segue limitado por custo de encoding em tempo real e por aceleração de
hardware desigual; Safari não expõe encoding AV1 pelo WebRTC.

No lado do áudio, **Opus** é o padrão universal; o **Lyra** do Google opera na faixa de
3 kbps e é usado em cenários de banda muito baixa (Duo), não como codec principal de
conferência.

## SFU em cascata

Um SFU único é um ponto de concentração: todo mundo conecta na mesma máquina, e quem
está longe dela paga a latência inteira.

**Octo**, do Jitsi (2018), foi o primeiro desenho público: vários videobridges
interligados, encapsulando RTP num cabeçalho fixo simples e trocando também mensagens
de texto. Cada participante conecta no bridge mais próximo; os bridges trocam mídia
entre si. Remove o teto de participantes por conferência e baixa o atraso fim a fim
para grupos geograficamente espalhados.

**LiveKit** empacotou a mesma ideia como comportamento padrão: uma sessão deixa de ser
um objeto físico numa máquina e vira um objeto lógico que se espalha por nós e
datacenters, e o roteamento entre nós é do cluster.

Vale notar o que isso é: **distribuição de streaming sem P2P**. A árvore existe, mas os
nós internos são servidores de confiança em vez de espectadores. Compre isso ou compre
[[p2ptv-fundamentos]] — as duas coisas resolvem o mesmo problema de fan-out com
propriedades de confiança e custo opostas.

## Controle de congestionamento

Comum a todos: estimativa de banda por *transport-wide congestion control* (feedback
RTCP com chegada de cada pacote), pacotes de sondagem, e recuperação por NACK/RTX,
FEC e ocultação de perda. O SFU processa os receiver reports e, no desenho do Discord,
comunica ao emissor as camadas desejadas via *media sink wants*.
