# P2PTV — fundamentos

Distribuição de vídeo ao vivo em que os próprios espectadores retransmitem o conteúdo.
O servidor deixa de pagar `N × bitrate` de egresso e passa a pagar uma fração dele; a
capacidade de upload cresce junto com a audiência em vez de contra ela.

O preço é uma restrição que download P2P de arquivo não tem: **o dado tem prazo**. Um
chunk que chega depois do instante de reprodução é lixo. Todo o resto do desenho —
topologia, escalonamento, incentivo — existe para respeitar esse prazo.

## As três restrições

- **Prazo de reprodução.** O buffer é a folga. Buffer grande esconde jitter e churn,
  mas vira atraso ao vivo. Sistemas clássicos (PPLive, SopCast) operavam com dezenas de
  segundos de atraso; qualquer coisa "conversacional" está fora de alcance aqui.
- **Assimetria de upload.** A maioria das linhas domésticas sobe muito menos do que
  desce. Se o bitrate for próximo do upload médio do enxame, o sistema não fecha conta:
  a razão upload disponível / bitrate é o teto de tudo.
- **Churn.** Espectador troca de canal. O par que era seu fornecedor some sem aviso, e
  a recuperação precisa acontecer dentro do buffer.

## Topologias de overlay

**Tree-push.** Uma ou mais árvores enraizadas na fonte; o pai empurra os chunks ao
filho. Atraso baixo e previsível, overhead de controle quase nulo. Frágil: a saída de
um nó interno derruba toda a subárvore, e folhas não contribuem upload. Multi-tree
(descrito em SNAP e afins) divide o fluxo em sub-streams, cada um com sua árvore, para
que todo par seja interno em pelo menos uma — recupera a contribuição das folhas ao
custo de manter k árvores.

**Mesh-pull.** Não há estrutura fixa. Cada par mantém vizinhos, troca **buffer maps**
(bitmap de quais chunks possui) e **puxa** o que falta de quem tem. Resistente a churn
por construção — a falta de um vizinho é só uma linha a menos no mapa. Custo: overhead
de sinalização e um atraso a mais por salto (anunciar, pedir, receber). Foi o que
venceu na prática — PPLive, PPStream e SopCast são mesh-pull.

**Híbrido push-pull.** Pull para descobrir e se recuperar, push para o regime
permanente: uma vez que A viu que B tem o fluxo, B empurra os chunks seguintes sem
esperar pedido, e o pull volta só quando há buraco. Elimina o round-trip por chunk sem
herdar a fragilidade da árvore. É o desenho recomendado pela literatura moderna e o que
SopCast aproxima com sua relação pai/filho.

## Escalonamento de chunks

O escalonador responde duas perguntas por rodada: **qual chunk** e **de qual par**.

- **Rarest first**, herdado do BitTorrent, maximiza diversidade e vazão agregada, mas
  ignora o prazo — ótimo para arquivo, ruim para ao vivo.
- **Deadline-aware / earliest-first** prioriza o que vai tocar primeiro. Ótimo para
  continuidade, péssimo para diversidade: todo mundo quer o mesmo chunk ao mesmo tempo.
- Na prática se usa uma mistura ponderada: urgência perto do ponto de reprodução,
  raridade na cauda do buffer.

Escolha de fornecedor pesa capacidade anunciada, RTT medido e localidade (mesmo ASN,
mesma região). Localidade não é enfeite: é o que evita que tráfego P2P atravesse
trânsito caro e é a variável que mais move o percentual de descarga.

## Incentivo e defesa

- **Free-riding.** Sem contrapartida, o par consome e não sobe. Tit-for-tat estilo
  BitTorrent funciona mal ao vivo (a janela útil é curta demais para negociar), então
  os sistemas usam choke/unchoke simples, cotas ou reputação.
- **Poluição de conteúdo.** Um par malicioso injeta chunk corrompido. Sem verificação
  criptográfica por chunk, o lixo se propaga. É exatamente o problema que o PPSPP
  resolve com árvore de Merkle — ver [[ppspp-rfc7574]].
- **Privacidade.** Todo par vê o endereço IP de quem o serve. Isso é intrínseco ao P2P
  e é a razão declarada pela qual o Discord se recusa a usá-lo — ver
  [[discord-arquitetura-de-midia]].

## Métricas que importam

Continuidade de reprodução (fração de chunks entregues no prazo), atraso da fonte até a
tela, tempo até o primeiro quadro, taxa de descarga (fração servida por pares) e
eficiência de upload. As duas últimas são o argumento econômico; as três primeiras são
a razão pela qual o argumento econômico pode não valer.

Segue em [[cdn-p2p-hibrido]] o que sobrou disso tudo na web moderna.
