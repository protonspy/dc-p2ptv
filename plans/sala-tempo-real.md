---
autonomy: auto
ci: wait
status: approved
checksum: 9eb3f0a8b3bb8e8cab27d21bc6385350ed776b2896c758557db6f72c0f431f85
---

# Sala em tempo real

Rede federada de transmissão compartilhada. Quem quiser sobe um nó, vira dono de uma
comunidade e cria salas nela; os demais nós ajudam a encaminhar a mídia sem conseguir
abri-la. Uma sala tem até oito participantes, que falam e podem abrir a tela, e dezenas de
espectadores, que só recebem — sempre sub-segundo.

## Why

Fan-out de vídeo em tempo real custa banda, e banda é o que uma comunidade tem de sobra e
um servidor central tem de caro. A aposta é que a rede seja construída por quem a usa: o
dono do nó hospeda a própria comunidade, e os outros nós encaminham a transmissão dele em
árvore, de modo que a origem envie uma cópia por nó vizinho em vez de uma por espectador.
Isso obriga a duas coisas desde a primeira linha — separar o plano de controle, que decide
quem entra e a que sala, do plano de dados, que apenas encaminha; e cifrar a mídia com uma
chave que o nó dono entrega ao grupo dele, de modo que nenhum nó de passagem consiga
abri-la. A iniciativa está pronta quando uma sala de oito participantes com dezenas de
espectadores roda pela rede, atravessando dois níveis de árvore, com latência vidro a vidro
sub-segundo medida, banda economizada medida contra a árvore de um nó só, e um nó de
passagem em nenhum momento consegue decodificar um quadro.

## Paths

- `server/` — Go: plano de controle, nó de encaminhamento, roteamento, árvore
- `client/` — TypeScript: sala, malha, captura, cifra de quadro
- `docs/adr/` — as decisões difíceis de reverter que este plano toma

## References

- `specs/comunidade-de-no/` — dono do nó, membros do grupo, o token que autoriza mudanças na rede, e a API REST que cria, lista e encerra sala
- `specs/sinalizacao/` — a sessão em WebSocket: presença, estado da sala, relay de SDP e ICE
- `specs/malha-direta/` — mídia navegador a navegador enquanto a sala tem até quatro participantes e nenhum espectador
- `specs/promocao-de-sala/` — a passagem da malha para a rede, e a de espectador a participante, sem cortar a chamada
- `specs/sfu-de-no/` — o nó de encaminhamento sobre Pion, opaco ao conteúdo que carrega
- `specs/transmissao/` — o papel de espectador dentro da sala, e o comando que o anfitrião tem sobre a transmissão
- `specs/camadas-simulcast/` — camadas de qualidade e a escolha de camada por assinante
- `specs/roteamento-de-nos/` — registro, saúde, afinidade da sala por hash, escolha de nó, remigração quando um nó some
- `specs/cascata-entre-nos/` — a árvore de distribuição entre nós, sua profundidade e seu reparo
- `specs/federacao-de-nos/` — identidade de nó, autenticação mútua, bilhete de sala, reputação
- `specs/chave-de-grupo/` — a chave que o nó dono gera, entrega ao grupo e rotaciona
- `specs/quadro-cifrado/` — cifra do quadro no cliente e o que fica em claro para o nó rotear
- `specs/cliente-de-sala/` — mídia local, dispositivos, entrada e saída, apresentação da sala
- `specs/compartilhamento-de-tela/` — captura de tela, dica de conteúdo, áudio do sistema
- `specs/metricas-de-sessao/` — latência vidro a vidro, tempo até o primeiro quadro, travamento, banda economizada

## Out of scope

- Espectador retransmitindo para espectador. A árvore vai até o nó, e para aí.
- Audiência de centenas numa transmissão, e a profundidade de árvore que ela exigiria.
- Difusão com segmentos e enxame no molde do PPSPP. É outro produto, com outra latência.
- Mais de oito participantes emitindo na mesma sala.
- Espectador de fora da comunidade do nó dono. A sala pertence à comunidade, e a chave não sai dela.
- Integração com o Discord, em qualquer forma.
- Gravação, catálogo e reprodução sob demanda.
- Transcodificação e faixas de bitrate geradas no servidor. O nó encaminha o que recebe.
- Cifra que esconda a mídia do próprio dono do nó. Ele é o administrador da comunidade dele, e isso é declarado a quem entra.
- Proteção de conteúdo por licença, no molde do Widevine ou do FairPlay.

## Tasks

- [x] 1.1 (Unit) Montar o repositório com os projetos Go e TypeScript, e registrar em `.claude/rules/project.md` os comandos de construção, teste, lint e formatação
- [x] 1.2 (Unit) Assentar em `docs/glossary.md` o vocabulário que os specs vão herdar: sala, conversa, transmissão, participante, espectador, nó, comunidade, malha, rede, promoção, camada, bilhete
- [x] 1.3 (Unit) Registrar em `docs/stack.md` cada tecnologia adotada, com a linha que diz por que ela entrou
  _Depends 1.1_
- [x] 1.4 (Unit) Registrar em ADR a separação entre plano de controle e plano de dados, e o que um nó de passagem nunca decide
- [x] 1.5 (Unit) Registrar em ADR a afinidade da sala por hash consistente e o teto de nós por sala, com o que cada um assume do outro
- [x] 1.6 (Unit) Registrar em ADR o modelo de confiança da cifra: chave do nó dono, opacidade para os nós de passagem, e o que o dono da comunidade consegue ver
- [x] 1.7 (Unit) Subir o serviço TURN sobre `pion/turn` com credenciais efêmeras, e instrumentar a fração de sessões que o usa
  _Depends 1.1_
- [x] 1.8 (Unit) Prover a chave de configuração que restringe a rede a nós próprios, para desligar a federação sem soltar versão
- [ ] 2.1 (Unit) Montar o ambiente de demonstração com três nós, uma sala de oito participantes e dezenas de espectadores
  _Depends 1.7_
- [ ] 2.2 (TDD) Medir latência vidro a vidro, tempo até o primeiro quadro e travamento, na malha e na rede, e comparar contra o piso sem promoção
  _Depends 2.1_
- [ ] 2.3 (Unit) Verificar, com um nó de passagem instrumentado para tentar decodificar o que encaminha, que nenhum quadro é legível fora do grupo
  _Depends 2.1_
- [ ] 2.4 (TDD) Medir a banda que a árvore economiza, comparando a saída do nó de origem contra a mesma transmissão servida por um nó só
  _Depends 2.1_
- [ ] 1.9 (Unit) Publicar em `docs/` o protocolo de fio — mensagens, sequência e formato
      — de modo que outro cliente o implemente sem ler o nosso
  _Reason o cliente deste plano é protótipo; o projeto discord-screen-p2p usará o mesmo servidor depois_
- [x] 1.10 (Unit) Registrar em ADR o transporte por faixas de mídia WebRTC, contra o
      WebCodecs sobre WebSocket que o projeto antecessor mede em 40 ms
  _Reason há alternativa em produção com número medido, e a escolha decide o cliente inteiro_
- [x] 1.11 (Unit) Prover integração contínua que rode construção, teste, lint e `scc
      validate` a cada pull request
  _Reason os comandos existiam e nada os rodava na abertura de uma pull request_
- [x] 1.12 (Unit) Exigir cobertura de teste igual ou acima de 86% nos dois projetos, e
      proteger `main` para que a mesclagem dependa dos checks
  _Reason pedido depois da aprovação: a integração contínua tinha de bloquear a mesclagem, não apenas relatar_

## Done when

- Uma sala de dois a quatro participantes, sem espectadores, troca vídeo e tela sem que nenhum byte de mídia atravesse um servidor, confirmado pelo par de candidatos ICE selecionado em cada conexão.
- Uma sala de cinco a oito participantes roda pela rede, e a promoção do quarto para o quinto acontece sem que a chamada congele.
- A tela de um participante alcança dezenas de espectadores atravessando dois níveis de árvore, e o que o nó de origem envia é uma cópia por nó vizinho.
- A banda medida na origem da transmissão é uma fração da que um nó só teria enviado.
- A latência vidro a vidro medida fica abaixo de um segundo para participante e para espectador.
- Um nó de passagem tentando decodificar o que encaminha não obtém quadro algum.
- Uma sala é criada, listada e encerrada pela API REST do nó que a hospeda.
- Um nono participante é recusado, e um espectador de outra comunidade também.
- `scc validate` sai com zero, e todos os specs referenciados estão fechados.
