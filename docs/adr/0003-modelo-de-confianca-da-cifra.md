---
status: accepted
---

# 0003 — Modelo de confiança da cifra

## Context

A mídia de uma sala atravessa nós que a comunidade não escolheu. Sem cifra, quem hospeda
qualquer nó da rede assiste — e grava — qualquer transmissão que passe por ele.

Havia duas respostas possíveis, com custos muito diferentes.

Cifra fim a fim **entre os participantes**, no molde do MLS que o Discord usa em DAVE:
nenhum servidor abre a mídia, nem o nó de passagem, nem o nó dono, nem quem opera a
rede. É a garantia mais forte, e traz grupo, epochs, rekey a cada entrada e saída, e
verificação de identidade entre participantes.

Cifra com **chave de grupo do nó dono**: quem hospeda a comunidade gera a chave e a
entrega aos membros. Os nós de passagem encaminham bytes opacos.

A escolha depende de quem precisa ser protegido de quem. O dono do nó já é o
administrador daquela comunidade — cria as salas, define quem é membro, remove quem
abusa. Ele ocupa a mesma posição de quem opera um servidor de Matrix. Esconder a mídia
dele contradiz o papel que ele exerce, e custa o spec mais difícil do projeto.

## Decision

A mídia é cifrada com uma **chave de grupo gerada pelo nó dono** e entregue aos membros
da comunidade pelo plano de controle. A chave avança de geração a cada mudança de
participante na sala, e o quadro carrega a geração a que pertence.

Um **nó de passagem nunca abre a mídia**: recebe carga cifrada e o cabeçalho de que
precisa para rotear e escolher camada.

O **dono do nó consegue assistir** às salas da comunidade dele. Isso é declarado a quem
entra numa sala, na interface, e não em nota de rodapé.

A audiência de uma sala é limitada à comunidade do nó dono, de modo que a chave nunca
circula fora do grupo que aquele dono conhece.

## Consequences

O spec de cifra é um spec de tamanho comum, e não o mais difícil do projeto. Não há
grupo MLS, epochs, nem árvore de ratchet a implementar.

A garantia que a rede oferece muda de "ninguém vê" para "nenhum nó de passagem vê, e o
administrador da sua comunidade vê". Prometer mais do que isso seria falso, e a promessa
correta precisa aparecer onde a pessoa entra.

Quem quiser esconder a mídia do próprio administrador não é atendido por este desenho.
Se isso passar a ser exigido, é um registro novo que substitui este, não uma emenda.

Rotação a cada mudança de participante impede que quem saiu continue decifrando o que
vem depois. É o que torna "remover alguém da sala" um ato com efeito, e não apenas uma
mudança de lista.
